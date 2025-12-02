// ! MARK: 型定義
/** 
 * @typedef {[number,number,number]} Position3D
 * @typedef {[number,number]} Position2D
 * @typedef {{[alias: string]: Texture}} Textures
 */
/**
 * @typedef {object} UV
 * @property {string} alias 割り当てテクスチャ
 * @property {Position2D} start テクスチャの始点
 * @property {Position2D} end テクスチャの終点
 */
/**
 * @typedef {object} Texture
 * @property {WebGLTexture} texture
 * @property {number} height
 * @property {number} width
 */

// * MARK: deg2Rad()
/**
 * 角度をラジアンに変換する
 * @param {number} degrees 角度
 * @returns {number} ラジアン
 */
function deg2Rad(degrees) {
  return degrees * (Math.PI / 180)
}

// * MARK: loadImages()
/**
 * テクスチャを一括読み込みする関数
 * @param {WebGLRenderingContext} gl
 * @param {{[alias: string]: string}} assets - { alias: "path/to/image.png" } の形式
 * @returns {Promise<Textures>} 読み込まれたテクスチャの辞書オブジェクト
 */
async function loadTextures(gl, assets) {
  // 1. 各テクスチャの読み込みPromiseを作成
  const promises = Object.entries(assets).map(([alias, src]) => {
    return new Promise((resolve, reject) => {
      const img = new Image();

      img.onload = function () {
        // テクスチャオブジェクトの作成
        const texture = loadImageTexture(gl)
        if (!texture) {
          reject(new Error(`テクスチャの作成に失敗しました: ${alias}`));
          return;
        }

        // 画像データをGPUに転送
        gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
        // 成功時に alias と texture, 画像サイズを返す
        resolve({ alias, texture, width: img.width, height: img.height });
      };

      img.onerror = function () {
        reject(new Error(`画像の読み込みに失敗しました: ${src} (alias: ${alias})`));
      };

      img.src = src;
    });
  });

  try {
    const loadedResults = await Promise.all(promises);

    // 読み込み結果からテクスチャマップとサイズマップを生成
    /** @type {Textures} */
    const textureMap = {};

    loadedResults.forEach((item) => {
      // item は { alias, texture, width, height } を含む
      textureMap[item.alias] = {
        texture: item.texture,
        height: item.height,
        width: item.width
      }
    });
    // テクスチャマップとサイズマップを同時に返す
    return textureMap;
  } catch (error) {
    throw error;
  }
}


// * MARK: loadImageTexture()
/**
 * 
 * @param {WebGLRenderingContext} gl 
 * @returns 
 */
function loadImageTexture(gl) {
  const texture = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, texture);

  // 画像が読み込まれるまでの仮のテクスチャ（1x1ピクセルの白）
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE,
    new Uint8Array([255, 255, 255, 255]));

  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);

  return texture;
}

// * MARK: makePerspective()
/**
 * 透視投影行列を作成 (列優先)
 * @param {number} fov 視野角 (ラジアン)
 * @param {number} aspect アスペクト比
 * @param {number} near 近平面
 * @param {number} far 遠平面
 * @returns {Float32Array}
 */
function makePerspective(fov, aspect, near, far) {
  const f = 1.0 / Math.tan(fov / 2);
  const nf = 1 / (near - far);
  return new Float32Array([
    f / aspect, 0, 0, 0,
    0, f, 0, 0,
    0, 0, (far + near) * nf, -1,
    0, 0, (2 * far * near) * nf, 0
  ]);
}

// * MARK: makeTranslation()
/**
 * 平行移動行列を作成
 * @param {number} tx 
 * @param {number} ty 
 * @param {number} tz 
 * @returns {Float32Array}
 */
function makeTranslation(tx, ty, tz) {
  return new Float32Array([
    1, 0, 0, 0,
    0, 1, 0, 0,
    0, 0, 1, 0,
    tx, ty, tz, 1
  ]);
}

// * MARK: makeRotationX()
/**
 * X軸回転行列を作成 (列優先)
 * @param {number} angleRad 
 * @returns {Float32Array}
 */
function makeRotationX(angleRad) {
  const c = Math.cos(angleRad);
  const s = Math.sin(angleRad);
  return new Float32Array([
    1, 0, 0, 0,
    0, c, s, 0,
    0, -s, c, 0,
    0, 0, 0, 1
  ]);
}
// * MARK: makeRotationY()
/**
 * Y軸回転行列を作成 (列優先)
 * @param {number} angleRad 
 * @returns {Float32Array}
 */
function makeRotationY(angleRad) {
  const c = Math.cos(angleRad);
  const s = Math.sin(angleRad);
  return new Float32Array([
    c, 0, -s, 0,
    0, 1, 0, 0,
    s, 0, c, 0,
    0, 0, 0, 1
  ]);
}

/**
 * Z軸回転行列を作成 (列優先)
 * @param {number} angleRad 
 * @returns {Float32Array}
 */
function makeRotationZ(angleRad) {
  const c = Math.cos(angleRad);
  const s = Math.sin(angleRad);
  return new Float32Array([
    c, s, 0, 0,
    -s, c, 0, 0,
    0, 0, 1, 0,
    0, 0, 0, 1
  ]);
}

// * MARK: multiplyMatrix()
/**
 * 4x4 行列の掛け算 (結果 = A * B, Aが左, Bが右)
 * WebGLは列優先なので注意。
 * @param {Float32Array} a 左側の行列
 * @param {Float32Array} b 右側の行列
 * @returns {Float32Array}
 */
function multiplyMatrix(a, b) {
  const out = new Float32Array(16);
  for (let i = 0; i < 4; i++) {
    for (let j = 0; j < 4; j++) {
      let sum = 0;
      for (let k = 0; k < 4; k++) {
        sum += a[i + k * 4] * b[k + j * 4];
      }
      out[i + j * 4] = sum;
    }
  }
  return out;
}

/**
 * 
 * @param {[number,number,number]} V vertex position
 * @property {[number,number,number]} center 回転中心
 * @property {[number,number,number]} angle 回転角度[rad]
 * @returns {[number,number,number]}
 */
function calcTransformation(V, center, angle) {
  let [x, y, z] = V
  const [cx, cy, cz] = center
  const [rx, ry, rz] = angle

  // 1. 回転中心に移動
  // V' = V - C
  x -= cx
  y -= cy
  z -= cz
  // 2. 回転 (X-Y-Z)
  // 2-1. X軸回転 (pitch)
  if (rx !== 0) {
    const cosX = Math.cos(rx)
    const sinX = Math.sin(rx)
    const newY = y * cosX - z * sinX
    const newZ = y * sinX + z * cosX
    y = newY
    z = newZ
  }
  // 2-2. Y軸回転 (yaw)
  if (ry !== 0) {
    const cosY = Math.cos(ry)
    const sinY = Math.sin(ry)
    const newX = x * cosY + z * sinY
    const newZ = -x * sinY + z * cosY
    x = newX
    z = newZ
  }
  // 2-3. Z軸回転 (roll)
  if (rz !== 0) {
    const cosZ = Math.cos(rz)
    const sinZ = Math.sin(rz)
    const newX = x * cosZ - y * sinZ
    const newY = x * sinZ + y * cosZ
    x = newX
    y = newY
  }

  return [x, y, z]
}

// ! MARK: [CLASS] Square 
/**
 * @typedef SquareOptions
 * @property {string[]} [tag] 識別子
 * @property {[number,number]} length (0,0)からの辺長さ(左上,右下])
 * @property {SquareTransformation} base (0,0)からどのぐらい回すのか[rad]
 * @property {SquareTransformation} [effect] 基本回転・移動後からどのようにするか
 * @property {UV} uv 使用テクスチャ/UV
 */
/**
 * @typedef {object} SquareTransformation
 * @property {[number,number,number]} center 回転中心
 * @property {[number,number,number]} angle 回転角度[rad]
 * @property {[number,number,number]} offset オフセット
 */
/**
 * @typedef {object} Face
 * @property {string} uvAliasName
 * @property {number[]} vertexPositions
 * @property {number[]} textureCoordinates
 */

/**
 * WebGLで長方形の面を取り扱えるようにする
 */
class Square {
  /** @type {string[]?} 識別子*/
  tag
  /** @type {[number,number]} (0,0)からの面頂点(左上,右下]) */
  length
  /** @type {SquareTransformation} */
  base
  /** @type {SquareTransformation?} */
  effect
  /** @type {UV} 使用テクスチャ/UV */
  uv

  // * MARK: > constructor()
  /**
   * @param {SquareOptions} options
   */
  constructor(options) {
    const {
      tag = [],
      length = [1, 1],
      base = {
        center: [0, 0, 0],
        angle: [0, 0, 0],
        offset: [0, 0, 0],
      },
      effect = {
        center: [0, 0, 0],
        angle: [0, 0, 0],
        offset: [0, 0, 0],
      },
      uv = { alias: "null", start: [0, 0], end: [0, 0] }
    } = options;

    // すべてのプロパティを初期化
    this.tag = tag
    this.length = length;
    this.base = base;
    this.effect = effect;
    this.uv = uv;
  }

  // * MARK: > toWebGL()
  /**
   * @param {Textures} textures
   * @returns {Face}
   */
  toWebGL(textures) {
    const { length, uv } = this

    const center = [length[0] / 2, length[1] / 2]
    // 頂点の定義
    const V1 = [-center[0], center[1], 0] // 左上
    const V2 = [center[0], center[1], 0] // 右上
    const V3 = [center[0], -center[1], 0] // 右下
    const V4 = [-center[0], -center[1], 0] // 左下
    // 頂点座標の回転・平行移動
    const TV1 = this.applyTransformation(V1)
    const TV2 = this.applyTransformation(V2)
    const TV3 = this.applyTransformation(V3)
    const TV4 = this.applyTransformation(V4)

    // 2つの三角形に分解
    const vertexPositions = [
      ...TV1, ...TV2, ...TV3,
      ...TV3, ...TV4, ...TV1
    ];

    // テクスチャ座標の正規化
    const [u_px_start, v_px_start] = uv.start;
    const [u_px_end, v_px_end] = uv.end;

    const width = textures[uv.alias].width
    const height = textures[uv.alias].height
    const u_start = u_px_start / width;
    const v_start = v_px_start / height;
    const u_end = u_px_end / width;
    const v_end = v_px_end / height;

    const UV_TOP_LEFT = [u_start, v_start];
    const UV_TOP_RIGHT = [u_end, v_start];
    const UV_BOTTOM_RIGHT = [u_end, v_end];
    const UV_BOTTOM_LEFT = [u_start, v_end];

    const textureCoordinates = [
      ...UV_TOP_LEFT,
      ...UV_TOP_RIGHT,
      ...UV_BOTTOM_RIGHT,
      ...UV_BOTTOM_RIGHT,
      ...UV_BOTTOM_LEFT,
      ...UV_TOP_LEFT
    ];

    return {
      uvAliasName: uv.alias,
      vertexPositions: vertexPositions,
      textureCoordinates: textureCoordinates
    }
  }

  // * MARK: > applyTransfrom()
  /**
   * 回転用中心に移動 => 回転 => オフセット
   * @param {[number,number,number]} V
   */
  applyTransformation(V) {
    // 1. モデル描画用の基礎回転
    const [bx, by, bz] = calcTransformation(V, this.base.center, this.base.angle)
    // 2. アニメーション用の基礎回転
    let [x, y, z] = calcTransformation([bx,by,bz], this.effect.center, this.effect.angle)
    // 3. オフセットによる移動 (平行移動)
    x += this.base.offset[0] + this.effect.offset[0]
    y += this.base.offset[1] + this.effect.offset[1]
    z += this.base.offset[2] + this.effect.offset[2]

    return [x, y, z]
  }
}

// ! MARK: [CLASS] Camera
class Camera {
  /** @type {[number,number,number]} カメラの位置 (ワールド座標) */
  offset;
  /** @type {[number,number,number]} カメラの回転角度 (ラジアン: [X軸回転, Y軸回転, Z軸回転]) */
  rotation;
  /** @type {Float32Array} 透視投影行列 (Perspective Matrix) */
  projectionMatrix;

  /**
   * @param {object} [options]
   * @param {[number,number,number]} [options.offset=[0,0,0]] カメラの位置
   * @param {[number,number,number]} [options.rotation=[0,0,0]] 回転角度 (ラジアン)
   */
  constructor(options = {}) {
    const {
      offset = [0, 0, 0],
      rotation = [0, 0, 0] // [X, Y, Z]
    } = options;

    this.offset = offset;
    this.rotation = rotation;
    // 初期状態では単位行列 (Identity Matrix)
    this.projectionMatrix = new Float32Array([
      1, 0, 0, 0,
      0, 1, 0, 0,
      0, 0, 1, 0,
      0, 0, 0, 1
    ]);
  }

  /**
     * 透視投影行列を計算し、内部に保持します。
     * @param {number} fovRad 垂直視野角 (ラジアン)
     * @param {number} aspect アスペクト比 (幅/高さ)
     * @param {number} near 手前のクリッピング面距離 (0より大きい値)
     * @param {number} far 奥のクリッピング面距離
     */
  setPerspective(fovRad, aspect, near, far) {
    this.projectionMatrix = makePerspective(fovRad, aspect, near, far);
  }

  /**
   * カメラのビュー行列 (ワールド座標からカメラ座標への変換) を計算して取得します。
   * @returns {Float32Array} 4x4 ビュー行列 (列優先)
   */
  getViewMatrix() {
    // 1. 逆回転 (R_c^{-1} を計算)
    const [rx, ry, rz] = this.rotation;
    let invRotationMatrix = makeRotationX(-rx);
    invRotationMatrix = multiplyMatrix(makeRotationY(-ry), invRotationMatrix);
    invRotationMatrix = multiplyMatrix(makeRotationZ(-rz), invRotationMatrix);

    // 2. 平行移動
    const TranslationMatrix = makeTranslation(this.offset[0], this.offset[1], this.offset[2]);

    // 3. 行列乗算: T_c^{-1} * R_c^{-1} 
    // T_c^{-1} (invTranslationMatrix) を左、 R_c^{-1} (invRotationMatrix) を右
    return multiplyMatrix(TranslationMatrix, invRotationMatrix);
  }

  /**
  * 投影行列とビュー行列を乗算し、最終的な PVM (Projection * View) 行列を取得します。
  * この行列をシェーダーに渡して頂点を変換します。
  * PVM行列 = Projection Matrix * View Matrix
  * @returns {Float32Array} 4x4 変換行列 (列優先)
  */
  getProjectionViewMatrix() {
    const viewMatrix = this.getViewMatrix();
    // 行列乗算: Projection * View
    return multiplyMatrix(this.projectionMatrix, viewMatrix);
  }
}