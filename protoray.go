//go:build ignore

package main

import (
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/Konstantin8105/vl"
	"github.com/gdamore/tcell/v2"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	runeStart = rune(byte(32))
	runeEnd   = rune('■')
)

func init() {
	runtime.LockOSThread()
}

var WindowRatio float32 = 0.6

func main() {
	// initialize
	var root vl.Widget
	var action chan func()

	// vl demo
	root, action = vl.Demo()

	// unicode table
	//	{
	//		var t vl.Text
	//		var str string
	//		for i := runeStart; i < runeEnd; i++ {
	//			str += " " + string(rune(i))
	//		}
	//		t.SetText(str)
	//		var sc vl.Scroll
	//		sc.Root = &t
	//		root = &sc
	//	}

	// run vl widget in OpenGL
	err := Run(root, action)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
}

func Run(root vl.Widget, action chan func()) (err error) {
	var (
		screenWidth  int32 = 800
		screenHeight int32 = 600
	)

	// TODO rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagVsyncHint)
	rl.InitWindow(screenWidth, screenHeight, "3D model")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second

	fontSize := float32(13)
	font := rl.LoadFontEx("./ProggyClean.ttf", int32(fontSize), nil)
	// font := rl.LoadFontEx("/home/konstantin/.fonts/Go-Mono.ttf", int32(fontSize), nil)

	// Generate mipmap levels to use trilinear filtering
	// NOTE: On 2D drawing it won't be noticeable, it looks like FILTER_BILINEAR
	rl.GenTextureMipmaps(&font.Texture)
	rl.SetTextureFilter(font.Texture, rl.FilterPoint)

	gw, gh := float32(font.Chars.AdvanceX+1), float32(fontSize)

	//mutex
	// 	var mutex sync.Mutex

	defaultColor := tcell.ColorWhite

	color := func(c tcell.Color) (cr rl.Color) {
		switch c {
		case tcell.ColorWhite:
			cr = rl.White
		case tcell.ColorBlack:
			cr = rl.Black
		case tcell.ColorRed:
			cr = rl.Red
		case tcell.ColorYellow:
			cr = rl.Yellow
		case tcell.ColorViolet:
			cr = rl.Blue
		case tcell.ColorMaroon:
			cr = rl.Orange
		default:
			panic(c)
		}
		return
	}

	// DrawText text on the screen
	DrawText := func(cell vl.Cell, x, y float32) {
		// We need to offset each string by the height of the
		// font. To ensure they don't overlap each other.
		x *= gw
		y *= gh

		// prepare colors
		fg, bg, attr := cell.S.Decompose()
		_ = attr

		if bg != defaultColor {
			rl.DrawRectangle(int32(x), int32(y), int32(gw), int32(gh), color(bg))
		}
		r := cell.R
		if r != 32 {
			rl.DrawTextEx(
				font,
				string(r),
				rl.Vector2{x, y},
				fontSize,
				0,
				color(fg),
			)
		}
	}

	screen := vl.Screen{
		Root: root,
	}
	var cells [][]vl.Cell

	var widthSymbol uint
	var heightSymbol uint
	var w, h, x int32

	//	window.SetMouseButtonCallback(func(
	//		w *glfw.Window,
	//		button glfw.MouseButton,
	//		action glfw.Action,
	//		mods glfw.ModifierKey,
	//	) {
	//		//mutex
	//		mutex.Lock()
	//		defer mutex.Unlock()
	//		//action
	//
	//		// convert button
	//		var bm tcell.ButtonMask
	//		switch button {
	//		case glfw.MouseButton1:
	//			bm = tcell.Button1
	//		case glfw.MouseButton2:
	//			bm = tcell.Button2
	//		case glfw.MouseButton3:
	//			bm = tcell.Button3
	//		default:
	//			return
	//		}
	//		// calculate position
	//		x, y := w.GetCursorPos()
	//		xs := int(x / float64(gw))
	//		ys := int(y / float64(gh))
	//		// create event
	//		switch action {
	//		case glfw.Press:
	//			screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
	//		case glfw.Release:
	//
	//		default:
	//			// case glfw.Repeat:
	//			// do nothing
	//		}
	//	})

	// Define the camera to look into our 3d world
	var camera rl.Camera3D
	camera.Position = rl.Vector3{10.0, 10.0, 10.0} // Camera position
	camera.Target = rl.Vector3{0.0, 0.0, 0.0}      // Camera looking at point
	camera.Up = rl.Vector3{0.0, 1.0, 0.0}          // Camera up vector (rotation towards target)
	camera.Fovy = 25.0                             // Camera field-of-view Y
	// camera.Projection = rl.CameraPerspective       // Camera mode type
	camera.Projection = rl.CameraOrthographic // Camera mode type

	cubePosition := rl.Vector3{0.0, 0.0, 0.0}

	// rl.SetCameraMode(camera, rl.CameraThirdPerson)
	rl.SetCameraMode(camera, rl.CameraOrbital)
	// rl.SetCameraMode(camera, rl.CameraFree) // Set a free camera mode

	screenGui := rl.LoadRenderTexture(int32(float32(screenWidth)*WindowRatio), screenHeight)
	defer rl.UnloadRenderTexture(screenGui) // Unload render texture
	screen3d := rl.LoadRenderTexture(int32(float32(screenWidth)*(1-WindowRatio)), screenHeight)
	defer rl.UnloadRenderTexture(screen3d) // Unload render texture

	for !rl.WindowShouldClose() {
		// windows
		w = int32(rl.GetScreenWidth())
		h = int32(rl.GetScreenHeight())
		x = int32(float32(rl.GetScreenWidth()) * WindowRatio)

		rl.UpdateCamera(&camera)

		// rc := color(defaultColor)
		// rl.ClearBackground(rc)

		{ // draw 3D model
			// Draw Player1 view to the render texture
			rl.BeginTextureMode(screen3d)

			rl.ClearBackground(rl.White)

			rl.BeginMode3D(camera)

			rl.DrawCube(cubePosition, 2.0, 2.0, 2.0, rl.Red)
			rl.DrawCubeWires(cubePosition, 2.0, 2.0, 2.0, rl.Maroon)
			rl.DrawCube(cubePosition, 9.0, 1.0, 1.0, rl.Red)
			rl.DrawCubeWires(cubePosition, 9.0, 1.0, 1.0, rl.Maroon)
			rl.DrawGrid(10, 1.0)

			rl.Begin(rl.RL_TRIANGLES)
			rl.Color3f(0.8, 0.8, 0.8)
			rl.Vertex3f(1, 0, 0)
			rl.Vertex3f(2, 2, 2)
			rl.Vertex3f(3, 4, 0)
			rl.End()

			rl.Begin(rl.RL_TRIANGLES)
			rl.Color3f(0.8, 0.8, 0.8)
			rl.Vertex3f(2, 2, 2)
			rl.Vertex3f(1, 0, 0)
			rl.Vertex3f(3, 4, 0)
			rl.End()

			rl.Begin(rl.RL_LINES)
			rl.Color3f(0.2, 0.2, 0.8)
			rl.Vertex3f(1, 0, 0)
			rl.Vertex3f(3, 4, 0)
			rl.End()

			rl.Begin(rl.RL_LINES)
			rl.Color3f(0.8, 0.2, 0.8)
			rl.Vertex3f(3, 4, 0)
			rl.Vertex3f(2, 2, 2)
			rl.End()

			rl.Begin(rl.RL_LINES)
			rl.Color3f(0.2, 0.8, 0.8)
			rl.Vertex3f(2, 2, 2)
			rl.Vertex3f(1, 0, 0)
			rl.End()

			// draw gizmo
			DrawGizmo(rl.Vector3{X: 2, Y: 2, Z: 2})

			DrawSpiral()

			rl.EndMode3D()
			rl.EndTextureMode()
		}

		{ // draw tui
			rl.BeginTextureMode(screenGui)
			rl.ClearBackground(rl.SkyBlue)
			rl.BeginDrawing()

			widthSymbol = uint(float32(w) / float32(gw) * WindowRatio)
			heightSymbol = uint(float32(h) / float32(gh))
			screen.SetHeight(heightSymbol)
			screen.GetContents(widthSymbol, &cells)
			for r := 0; r < len(cells); r++ {
				if len(cells[r]) == 0 {
					continue
				}
				for c := 0; c < len(cells[r]); c++ {
					DrawText(cells[r][c], float32(c), float32(r))
				}
			}

			rl.DrawFPS(10, 10)

			rl.EndDrawing()
			rl.EndTextureMode()
		}

		{
			// Draw both views render textures to the screen side by side
			rl.BeginDrawing()
			rl.ClearBackground(rl.Black)
			rl.DrawTextureRec(screenGui.Texture,
				rl.Rectangle{0.0, 0.0, float32(screenGui.Texture.Width), float32(-screenGui.Texture.Height)},
				rl.Vector2{0, 0}, rl.White)
			rl.DrawTextureRec(screen3d.Texture,
				rl.Rectangle{0.0, 0.0, float32(screen3d.Texture.Width), float32(-screen3d.Texture.Height)},
				rl.Vector2{float32(x), 0}, rl.White)
			rl.EndDrawing()
		}

		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			xs := int(float32(rl.GetMouseX()) / float32(gw))
			ys := int(float32(rl.GetMouseY()) / float32(gh))
			screen.Event(tcell.NewEventMouse(xs, ys, tcell.Button1, tcell.ModNone))
		}

		if key := rl.GetKeyPressed(); key >= 32 && key <= 126 {
			// TODO : is key inside tui? is it important.
			screen.Event(tcell.NewEventKey(tcell.KeyRune, rune(key), tcell.ModNone))
		}

		if yoffset := rl.GetMouseWheelMove(); yoffset != 0.0 {
			// TODO : is key inside tui? is it important.
			xs := int(float32(rl.GetMouseX()) / float32(gw))
			ys := int(float32(rl.GetMouseY()) / float32(gh))

			var bm tcell.ButtonMask
			if yoffset < 0 {
				bm = tcell.WheelDown
			}
			if 0 < yoffset {
				bm = tcell.WheelUp
			}
			// TODO :
			// if xoffset < 0 {
			// 	bm = tcell.WheelLeft
			// }
			// if 0 < xoffset {
			// 	bm = tcell.WheelRight
			// }
			screen.Event(tcell.NewEventMouse(xs, ys, bm, tcell.ModNone))
		}

	}

	return
}

// Draw gizmo
func DrawGizmo(position rl.Vector3) {
	// NOTE: RGB = XYZ
	var lenght float32 = 1.0

	// rl.PushMatrix()
	// rl.Translatef(position.X, position.Y, position.Z)
	// rlRotatef(rotation, 0, 1, 0);
	rl.Scalef(lenght, lenght, lenght)

	x, y, z := position.X, position.Y, position.Z

	rl.Begin(rl.RL_LINES)
	rl.Color3f(1.0, 0.0, 0.0)
	rl.Vertex3f(x, y, z)
	rl.Color3f(1.0, 0.0, 0.0)
	rl.Vertex3f(x+1.0, y, z)

	rl.Color3f(0.0, 1.0, 0.0)
	rl.Vertex3f(x, y, z)
	rl.Color3f(0.0, 1.0, 0.0)
	rl.Vertex3f(x, y+1, z)

	rl.Color3f(0.0, 0.0, 1.0)
	rl.Vertex3f(x, y, z)
	rl.Color3f(0.0, 0.0, 1.0)
	rl.Vertex3f(x, y, z+1)
	rl.End()
	//	rl.PopMatrix()
}

func DrawSpiral() {
	var (
		Ri     = 0.5
		Ro     = 2.5
		dR     = 0.0
		da     = 30.0 // degree
		dy     = 0.2
		levels = 60
		//    8 = FPS 61.0
		//   80 = FPS 58.0
		//  800 = FPS 25.0
		// 8000 = FPS  5.5 --- 16000 points
	)
	for i := 0; i < int(levels); i++ {
		Ro += dR
		Ri += dR
		angle := float64(i) * da * math.Pi / 180.0

		bc0 := float32(Ri * math.Sin(angle))
		bc1 := float32(float64(i) * dy)
		bc2 := float32(Ri * math.Cos(angle))

		// TODO: raylib  have not RL_POINTS
		// rl.Begin(rl.RL_POINTS)
		// rl.Color3f(0.8, 0.8, 0.8)
		// rl.Vertex3f(bc0, bc1, bc2)
		// rl.End()
		DrawPoint(bc0, bc1, bc2)

		fc0 := float32(Ro * math.Sin(angle))
		fc1 := float32(float64(i) * dy)
		fc2 := float32(Ro * math.Cos(angle))

		// TODO: raylib  have not RL_POINTS
		// rl.Begin(rl.RL_POINTS)
		// rl.Color3f(0.8, 0.1, 0.1)
		// rl.Vertex3f(fc0, fc1, fc2)
		// rl.End()
		DrawPoint(fc0, fc1, fc2)

		rl.Begin(rl.RL_LINES)
		rl.Color3f(0.3, 0.7, 0.1)
		rl.Vertex3f(bc0, bc1, bc2)
		rl.Vertex3f(fc0, fc1, fc2)
		rl.End()

	}
}

func DrawPoint(x, y, z float32) {
	var size float32 = 0.05
	rl.Color3f(0.3, 0.7, 0.8)
	rl.Begin(rl.RL_LINES)
	rl.Vertex3f(x+size, y, z)
	rl.Vertex3f(x-size, y, z)
	rl.End()
	rl.Begin(rl.RL_LINES)
	rl.Vertex3f(x, y+size, z)
	rl.Vertex3f(x, y-size, z)
	rl.End()
	rl.Begin(rl.RL_LINES)
	rl.Vertex3f(x, y, z+size)
	rl.Vertex3f(x, y, z-size)
	rl.End()
}
