package main

import (
	"fmt"
	"log"
	"regexp" // VID/PID抽出のためにregexpを追加
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Windows API のインポート
var (
	user32 = windows.NewLazySystemDLL("user32.dll")

	procCreateWindowExW         = user32.NewProc("CreateWindowExW")
	procDefWindowProcW          = user32.NewProc("DefWindowProcW")
	procGetMessageW             = user32.NewProc("GetMessageW")
	procTranslateMessage        = user32.NewProc("TranslateMessage")
	procDispatchMessageW        = user32.NewProc("DispatchMessageW")
	procRegisterClassExW        = user32.NewProc("RegisterClassExW")
	procRegisterRawInputDevices = user32.NewProc("RegisterRawInputDevices")
	procGetRawInputData         = user32.NewProc("GetRawInputData")
	// デバイス情報を取得するための関数を追加
	procGetRawInputDeviceInfoW = user32.NewProc("GetRawInputDeviceInfoW")
)

// 定義済み定数
const (
	WM_INPUT = 0x00FF

	RIM_TYPEMOUSE    = 0
	RIM_TYPEKEYBOARD = 1
	RIM_TYPEHID      = 2 // HIDデバイス (ジョイスティック、ゲームパッドなど)

	// GetRawInputDeviceInfoW のフラグ
	RIDI_DEVICENAME = 0x20000007 // デバイス名 (文字列) を取得

	// HID の Usage Page (0x01: Generic Desktop Controls)
	HID_USAGE_GENERIC_MOUSE    = 0x02
	HID_USAGE_GENERIC_KEYBOARD = 0x06
	HID_USAGE_GENERIC_JOYSTICK = 0x04
	HID_USAGE_GENERIC_GAMEPAD  = 0x05

	RIDEV_INPUTSINK = 0x00000100
	RID_INPUT       = 0x10000003

	// マウスボタンフラグ (usButtonFlags)
	RI_MOUSE_LEFT_BUTTON_DOWN   = 0x0001
	RI_MOUSE_LEFT_BUTTON_UP     = 0x0002
	RI_MOUSE_RIGHT_BUTTON_DOWN  = 0x0004
	RI_MOUSE_RIGHT_BUTTON_UP    = 0x0008
	RI_MOUSE_MIDDLE_BUTTON_DOWN = 0x0010
	RI_MOUSE_MIDDLE_BUTTON_UP   = 0x0020
	RI_MOUSE_WHEEL              = 0x0040 // 垂直ホイールスクロール (Z軸)
	RI_MOUSE_XBUTTON1_DOWN      = 0x0080 // XButton1 (サイドボタンなど) ダウン
	RI_MOUSE_XBUTTON1_UP        = 0x0100 // XButton1 アップ
	RI_MOUSE_XBUTTON2_DOWN      = 0x0200 // XButton2 ダウン
	RI_MOUSE_XBUTTON2_UP        = 0x0400 // XButton2 アップ
	RI_MOUSE_HWHEEL             = 0x0800 // 水平ホイールスクロール (X軸)
)

// 構造体定義

type MSG struct {
	Hwnd    windows.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

type WNDCLASSEX struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     windows.Handle
	hIcon         windows.Handle
	hCursor       windows.Handle
	hbrBackground windows.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       windows.Handle
}

type RAWINPUTDEVICE struct {
	usUsagePage uint16
	usUsage     uint16
	dwFlags     uint32
	hwndTarget  windows.Handle
}

type RAWINPUTHEADER struct {
	dwType  uint32
	dwSize  uint32
	hDevice windows.Handle // デバイスハンドル
	wParam  uintptr
}

// RAWMOUSE 構造体を修正: C言語の共用体をエミュレートするため ulButtons (uint32) を使用
type RAWMOUSE struct {
	usFlags uint16
	// ulButtons (uint32) is the Union field.
	ulButtons          uint32
	ulRawButtons       uint32 // 生のボタン状態
	lLastX             int32
	lLastY             int32
	ulExtraInformation uint32
}

type RAWKEYBOARD struct {
	MakeCode         uint16
	Flags            uint16
	Reserved         uint16
	VKey             uint16
	Message          uint32
	ExtraInformation uint32
}

type RAWINPUT struct {
	header RAWINPUTHEADER
	// data フィールドは可変長データのために用意されています
	data [1]byte
}

// GetModuleHandle の自前実装
func getModuleHandle(moduleName *uint16) windows.Handle {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	proc := kernel32.NewProc("GetModuleHandleW")
	h, _, _ := proc.Call(uintptr(unsafe.Pointer(moduleName)))
	return windows.Handle(h)
}

func utf16Ptr(s string) *uint16 {
	p, _ := windows.UTF16PtrFromString(s)
	return p
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// getDeviceName: デバイスハンドルからデバイス名（パス）を取得する
func getDeviceName(hDevice windows.Handle) string {
	var requiredSize uint32

	// 1. 必要なバッファサイズを取得
	ret, _, _ := procGetRawInputDeviceInfoW.Call(
		uintptr(hDevice),
		RIDI_DEVICENAME,
		0,
		uintptr(unsafe.Pointer(&requiredSize)),
	)

	if ret == ^uintptr(0) || requiredSize == 0 {
		return fmt.Sprintf("Unknown Device (Handle: 0x%X)", hDevice)
	}

	// 2. 実際のデバイス名を取得
	buf := make([]uint16, requiredSize)
	ret, _, _ = procGetRawInputDeviceInfoW.Call(
		uintptr(hDevice),
		RIDI_DEVICENAME,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&requiredSize)),
	)

	if ret == ^uintptr(0) {
		return fmt.Sprintf("Unknown Device (Handle: 0x%X, Error getting name)", hDevice)
	}

	name := syscall.UTF16ToString(buf)

	// VID/PIDを抽出するための正規表現 (例: VID_046D&PID_C08B)
	re := regexp.MustCompile(`VID_([0-9A-Fa-f]{4})&PID_([0-9A-Fa-f]{4})`)
	matches := re.FindStringSubmatch(name)

	if len(matches) >= 3 {
		return fmt.Sprintf("Device (VID: %s, PID: %s)", matches[1], matches[2])
	}

	return fmt.Sprintf("Device (Handle: 0x%X)", hDevice)
}

// WndProc (ウィンドウプロシージャ)
func wndProc(hwnd windows.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == WM_INPUT {
		var dataSize uint32

		// 1. サイズ取得
		ret, _, _ := procGetRawInputData.Call(
			lParam,
			RID_INPUT,
			0,
			uintptr(unsafe.Pointer(&dataSize)),
			unsafe.Sizeof(RAWINPUTHEADER{}),
		)
		if ret == 0 && dataSize == 0 {
			return 0
		}

		// 2. データ取得
		buf := make([]byte, dataSize)
		bytesCopied, _, _ := procGetRawInputData.Call(
			lParam,
			RID_INPUT,
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&dataSize)),
			unsafe.Sizeof(RAWINPUTHEADER{}),
		)
		if bytesCopied == 0 || bytesCopied == ^uintptr(0) {
			log.Printf("Error getting raw input data. Bytes copied: %d", bytesCopied)
			return 0
		}

		raw := (*RAWINPUT)(unsafe.Pointer(&buf[0]))
		deviceName := getDeviceName(raw.header.hDevice)

		switch raw.header.dwType {
		case RIM_TYPEKEYBOARD:
			// キーボード入力の処理
			headerSize := unsafe.Sizeof(RAWINPUTHEADER{})
			if dataSize < uint32(headerSize+unsafe.Sizeof(RAWKEYBOARD{})) {
				return 0
			}
			kb := (*RAWKEYBOARD)(unsafe.Pointer(&buf[headerSize]))

			action := "Down"
			if kb.Flags&0x01 != 0 {
				action = "Up"
			}

			fmt.Printf("[KB] Device: %s | VKey=0x%X (MakeCode: 0x%X, Action: %s)\n", deviceName, kb.VKey, kb.MakeCode, action)

		case RIM_TYPEMOUSE:
			// マウス入力の処理
			headerSize := unsafe.Sizeof(RAWINPUTHEADER{})
			if dataSize < uint32(headerSize+unsafe.Sizeof(RAWMOUSE{})) {
				return 0
			}
			mouse := (*RAWMOUSE)(unsafe.Pointer(&buf[headerSize]))

			// --- usButtonFlags と usButtonData の抽出 ---
			// usButtonFlags は ulButtons の下位16ビット (Low Word)
			usButtonFlags := uint16(mouse.ulButtons & 0xFFFF)
			// usButtonData は ulButtons の上位16ビット (High Word) (ホイールデルタ値)
			usButtonData := uint16(mouse.ulButtons >> 16)

			// usButtonFlags、usButtonData、lLastX/Y の全てがゼロの場合、メッセージをスキップする
			if usButtonFlags == 0 && usButtonData == 0 && mouse.lLastX == 0 && mouse.lLastY == 0 && mouse.ulRawButtons == 0 {
				ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
				return ret
			}

			// デバイス名と共に出力
			fmt.Printf("[Mouse] Device: %s | RawFlags: 0x%04X | RawButtons: 0x%04X | ulButtons: 0x%08X | ",
				deviceName, usButtonFlags, mouse.ulRawButtons, mouse.ulButtons)

			// --- ボタン状態のチェック ---

			var actionPrinted bool

			// ホイールスクロールの判定ロジックを修正
			isVWheel := (usButtonFlags&RI_MOUSE_WHEEL != 0) || (usButtonData != 0 && usButtonFlags&RI_MOUSE_HWHEEL == 0)
			isHWheel := usButtonFlags&RI_MOUSE_HWHEEL != 0

			// 垂直ホイールスクロール
			if isVWheel {
				delta := int16(usButtonData)
				// RI_MOUSE_WHEELフラグが立っていない場合でも、usButtonData (デルタ値) がゼロでなければホイールと見なす
				if delta != 0 {
					direction := "Up"
					if delta < 0 {
						direction = "Down"
					}
					fmt.Printf("V-Scroll %s (Delta: %d), ", direction, delta)
					actionPrinted = true
					// ホイールイベントとして処理したので、ボタンフラグからは除外する (表示上、ホイールだけをハイライトするため)
					usButtonFlags &^= RI_MOUSE_WHEEL
				}
			}

			// 水平ホイールスクロール
			if isHWheel {
				delta := int16(usButtonData)
				if delta != 0 {
					direction := "Right"
					if delta < 0 {
						direction = "Left"
					}
					fmt.Printf("H-Scroll %s (Delta: %d), ", direction, delta)
					actionPrinted = true
				}
			}

			// 残りのボタン状態のチェック (フラグが0x0040以外であることを確認)
			if usButtonFlags&RI_MOUSE_LEFT_BUTTON_DOWN != 0 {
				fmt.Print("LBUTTON Down, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_LEFT_BUTTON_UP != 0 {
				fmt.Print("LBUTTON Up, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_RIGHT_BUTTON_DOWN != 0 {
				fmt.Print("RBUTTON Down, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_RIGHT_BUTTON_UP != 0 {
				fmt.Print("RBUTTON Up, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_MIDDLE_BUTTON_DOWN != 0 {
				fmt.Print("MBUTTON Down, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_MIDDLE_BUTTON_UP != 0 {
				fmt.Print("MBUTTON Up, ")
				actionPrinted = true
			}
			// XButton (サイドボタンなど)
			if usButtonFlags&RI_MOUSE_XBUTTON1_DOWN != 0 {
				fmt.Print("XBUTTON1 Down, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_XBUTTON1_UP != 0 {
				fmt.Print("XBUTTON1 Up, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_XBUTTON2_DOWN != 0 {
				fmt.Print("XBUTTON2 Down, ")
				actionPrinted = true
			}
			if usButtonFlags&RI_MOUSE_XBUTTON2_UP != 0 {
				fmt.Print("XBUTTON2 Up, ")
				actionPrinted = true
			}

			// マウス移動
			if mouse.lLastX != 0 || mouse.lLastY != 0 {
				fmt.Printf("Move: dx=%d dy=%d (Flags: 0x%X)", mouse.lLastX, mouse.lLastY, mouse.usFlags)
				actionPrinted = true
			}

			// 何もアクションが検出されなかった場合
			if !actionPrinted {
				fmt.Print("No Action")
			}

			// 末尾に改行を追加
			fmt.Println()

		case RIM_TYPEHID:
			// HIDデバイスの処理
			fmt.Printf("[HID] Device: %s | Received raw HID input (Size: %d)\n", deviceName, dataSize)

		default:
			fmt.Printf("[Unknown] Device: %s | Unknown input type: %d (Size: %d)\n", deviceName, raw.header.dwType, dataSize)
		}
	}

	ret, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return ret
}

func main() {
	// 必要なハンドルとコールバックの設定
	hInstance := getModuleHandle(nil)

	className := utf16Ptr("DummyRawInputWindow")
	var wcex WNDCLASSEX
	wcex.cbSize = uint32(unsafe.Sizeof(wcex))
	// Windowsコールバック関数を登録
	wcex.lpfnWndProc = syscall.NewCallback(wndProc)
	wcex.hInstance = hInstance
	wcex.lpszClassName = className

	// ウィンドウクラスの登録
	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wcex)))
	if ret == 0 {
		log.Fatal("RegisterClassEx failed:", err)
	}

	// ダミーウィンドウの作成
	hwnd, _, err := procCreateWindowExW.Call(
		0,                                  // dwExStyle (WS_EX_...)
		uintptr(unsafe.Pointer(className)), // lpClassName
		uintptr(unsafe.Pointer(utf16Ptr("Raw Input Listener"))), // lpWindowName
		0,          // dwStyle (WS_...)
		0, 0, 0, 0, // x, y, nWidth, nHeight
		0, 0, // hWndParent, hMenu
		uintptr(hInstance), // hInstance
		0,                  // lpParam
	)
	if hwnd == 0 {
		log.Fatal("CreateWindowEx failed:", err)
	}

	// Raw Input 登録
	devices := []RAWINPUTDEVICE{
		{usUsagePage: 0x01, usUsage: HID_USAGE_GENERIC_MOUSE, dwFlags: RIDEV_INPUTSINK, hwndTarget: windows.Handle(hwnd)},    // マウス
		{usUsagePage: 0x01, usUsage: HID_USAGE_GENERIC_KEYBOARD, dwFlags: RIDEV_INPUTSINK, hwndTarget: windows.Handle(hwnd)}, // キーボード
		{usUsagePage: 0x01, usUsage: HID_USAGE_GENERIC_JOYSTICK, dwFlags: RIDEV_INPUTSINK, hwndTarget: windows.Handle(hwnd)}, // ジョイスティック
		{usUsagePage: 0x01, usUsage: HID_USAGE_GENERIC_GAMEPAD, dwFlags: RIDEV_INPUTSINK, hwndTarget: windows.Handle(hwnd)},  // ゲームパッド
	}
	ret, _, err = procRegisterRawInputDevices.Call(
		uintptr(unsafe.Pointer(&devices[0])),
		uintptr(len(devices)),
		unsafe.Sizeof(RAWINPUTDEVICE{}),
	)
	if ret == 0 {
		log.Fatal("RegisterRawInputDevices failed:", err)
	}

	fmt.Println("Raw input registered to window handle:", hwnd)
	fmt.Println("Running message loop. Input will be printed here even if the window is not focused.")

	// メッセージループ
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)

		if ret < 0 {
			log.Printf("GetMessageW failed with error code: %d", ret)
			break
		}
		if ret == 0 { // WM_QUIT
			break
		}

		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
	fmt.Println("Message loop finished.")
}
