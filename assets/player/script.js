//@ts-check
/// <reference path="./webgl.js" />
/// <reference path="./webgl_extend.js" />
/// <reference path="./const.js" />
/// <reference path="./websocket.js" />


// * MARK: Define Cubes
/** @type {Square[]} */
const minecraft_player_model = [
  // ...model.axis,
  ...model.heads,
  ...model.body,
  ...model.rightArmWide,
  // ...model.leftArmWideX32,
  ...model.leftArmWideX64,
  ...model.rightLeg,
  // ...model.leftLegX32,
  ...model.leftLegX64,
]

let animation_control = {
  walk: {
    isAnimation: false,
    timer: 0,
    scale: 0,
  },
  move: {
    front: false,
    back: false,
    right: false,
    left: false,
  },
  facing: {
    prevMousePos: [0, 0],
    nowMousePos: [0, 0],
    rotation: [0, 0],
    updateTimestamp: 0
  }
}

async function main() {
  // * MARK: Canvas Initialize
  /** @type {HTMLCanvasElement} */
  //@ts-expect-error
  const canvas = document.getElementById("canvas");

  /** @type {WebGLRenderingContext | null} */
  const gl = canvas.getContext("webgl");
  if (!gl) {
    alert("WebGLをサポートしていません");
    return;
  }

  // WebGLの初期設定
  gl.enable(gl.DEPTH_TEST);
  gl.viewport(0, 0, canvas.width, canvas.height);
  gl.clearColor(0.1, 0.1, 0.1, 1.0);

  // * MARK: WebGL Initialize
  // シェーダーのソースコード
  const vertexShaderSource = `
        attribute vec3 a_position;
        attribute vec2 a_texCoord;
        
        uniform mat4 u_matrix;
        
        varying vec2 v_texCoord;
        
        void main() {
          gl_Position = u_matrix * vec4(a_position, 1.0);
          v_texCoord = a_texCoord;
        }
      `;

  const fragmentShaderSource = `
        precision mediump float;
        
        varying vec2 v_texCoord;
        uniform sampler2D u_texture;
        
        void main() {
          gl_FragColor = texture2D(u_texture, v_texCoord);
        }
      `;

  // シェーダーとプログラムの作成
  const vertexShader = createShader(gl, gl.VERTEX_SHADER, vertexShaderSource);
  const fragmentShader = createShader(gl, gl.FRAGMENT_SHADER, fragmentShaderSource);
  if (!vertexShader || !fragmentShader) return;

  const program = createProgram(gl, vertexShader, fragmentShader);
  if (!program) return;

  // 画像が読み込まれたら描画開始
  const textures = await loadTextures(gl, {
    "skin": "./skin.png"
  })
  // アニメーション開始
  startAnimation(canvas.height / canvas.width, gl, program, textures);

  connectWebsocket()
}

/**
 * @param {number} aspectRatio
 * @param {WebGLRenderingContext} gl
 * @param {WebGLProgram} program
 * @param {Textures} textures
 */
function startAnimation(aspectRatio, gl, program, textures) {
  // バッファの作成
  const positionBuffer = gl.createBuffer();
  const texCoordBuffer = gl.createBuffer();
  // 属性の位置を取得
  const positionLocation = gl.getAttribLocation(program, "a_position");
  const texCoordLocation = gl.getAttribLocation(program, "a_texCoord");
  const matrixLocation = gl.getUniformLocation(program, "u_matrix");

  const camera = new Camera({
    offset: [0, 0, -50],
    rotation: [0, 0, 0]
  })
  camera.setPerspective(deg2Rad(45), aspectRatio, 0.1, 500);
  camera.rotation[0] = deg2Rad(45)
  camera.rotation[1] = deg2Rad(-160)
  camera.rotation[2] = deg2Rad(20)

  let rotation = 0
  function render() {
    rotation += 5;
    // あまたを自動で元に戻す
    if (animation_control.facing.updateTimestamp + (1000 * 5) < new Date().getTime()) {
      if (animation_control.facing.rotation[0] > 2) {
        animation_control.facing.rotation[0] -= 0.5
      }
      if (animation_control.facing.rotation[0] < -2) {
        animation_control.facing.rotation[0] += 0.5
      }
      if (animation_control.facing.rotation[1] > 2) {
        animation_control.facing.rotation[1] -= 0.5
      }
      if (animation_control.facing.rotation[1] < -2) {
        animation_control.facing.rotation[1] += 0.5
      }
    }
    // 手足の動き
    // 1.動かすべき?
    animation_control.walk.isAnimation = animation_control.move.front || animation_control.move.back || animation_control.move.right || animation_control.move.left
    // 2. 動き出し・動き終わりに合わせて手足の動きを抑える
    if (animation_control.walk.isAnimation) {
      if (animation_control.walk.scale < 20) {
        animation_control.walk.scale++
      }
    } else {
      if (animation_control.walk.scale > 0) {
        animation_control.walk.scale--
      } else {
        animation_control.walk.timer = 0
      }
    }
    const scale = animation_control.walk.scale / 20
    // 3. 動き中はtimerを増やしていく
    if (animation_control.walk.isAnimation && animation_control.walk.timer == 0) animation_control.walk.timer = 1
    if (!animation_control.walk.isAnimation && animation_control.walk.scale == 0) animation_control.walk.timer = 0
    if (animation_control.walk.timer > 0) {
      animation_control.walk.timer += 5
    }
    const rotate = Math.sin(deg2Rad(animation_control.walk.timer)) * 0.6 * scale

    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
    gl.useProgram(program);
    gl.uniformMatrix4fv(matrixLocation, false, camera.getProjectionViewMatrix());

    // 各面を描画
    minecraft_player_model.forEach(square => {
      if (square.tag?.includes("head")) {
        square.SetEffect({
          center: [0, 0, 0],
          angle: [deg2Rad(animation_control.facing.rotation[1]), deg2Rad(animation_control.facing.rotation[0]), 0],
          offset: [0, 0, 0],
        })
      }
      if (square.tag?.includes("rightArm")) {
        square.SetEffect({
          center: [0, 0, 0],
          angle: [rotate, deg2Rad(0), 0],
          offset: [0, 0, 0],
        })
      }
      if (square.tag?.includes("leftArm")) {
        square.SetEffect({
          center: [0, 0, 0],
          angle: [-rotate, deg2Rad(0), 0],
          offset: [0, 0, 0],
        })
      }
      if (square.tag?.includes("rightLeg")) {
        square.SetEffect({
          center: [0, 0, 0],
          angle: [-rotate * 1.25, deg2Rad(0), 0],
          offset: [0, 0, 0],
        })
      }
      if (square.tag?.includes("leftLeg")) {
        square.SetEffect({
          center: [0, 0, 0],
          angle: [rotate * 1.25, deg2Rad(0), 0],
          offset: [0, 0, 0],
        })
      }
      const data = square.ToWebGL(textures)
      // 頂点座標
      gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data.vertexPositions), gl.STATIC_DRAW);
      gl.enableVertexAttribArray(positionLocation);
      gl.vertexAttribPointer(positionLocation, 3, gl.FLOAT, false, 0, 0);

      // テクスチャ座標
      gl.bindBuffer(gl.ARRAY_BUFFER, texCoordBuffer);
      gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(data.textureCoordinates), gl.STATIC_DRAW);
      gl.enableVertexAttribArray(texCoordLocation);
      gl.vertexAttribPointer(texCoordLocation, 2, gl.FLOAT, false, 0, 0);

      // 描画
      gl.drawArrays(gl.TRIANGLES, 0, 6);
    })

    requestAnimationFrame(render);
  }
  render();
}

main();