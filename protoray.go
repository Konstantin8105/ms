//go:build ignore

package main

import (
	"fmt"
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

var WindowRatio float32 = 0.5

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

	rl.SetConfigFlags(rl.FlagWindowResizable | rl.FlagVsyncHint)
	rl.InitWindow(screenWidth, screenHeight, "3D model")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second

	font := rl.GetFontDefault()
	fontSize := rl.GetGlyphAtlasRec(font, '?')
	gw, gh := fontSize.Width, fontSize.Height

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
		rl.DrawText(string(cell.R), int32(x), int32(y), int32(gh), color(fg))

		// TODO implementation in raylib-go
		// rl.DrawTextCodepoint(font, int(cell.R), rl.Vector2{float32(x), float32(y)}, int32(gh), color(fg))
	}

	screen := vl.Screen{
		Root: root,
	}
	var cells [][]vl.Cell

	var widthSymbol uint
	var heightSymbol uint
	var w, h int32

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
	camera.Fovy = 45.0                             // Camera field-of-view Y
	camera.Projection = rl.CameraPerspective       // Camera mode type

	cubePosition := rl.Vector3{0.0, 0.0, 0.0}

	// rl.SetCameraMode(camera, rl.CameraThirdPerson)
	rl.SetCameraMode(camera, rl.CameraOrbital)
	// rl.SetCameraMode(camera, rl.CameraFree) // Set a free camera mode

	for !rl.WindowShouldClose() {
		// windows
		w = int32(rl.GetScreenWidth())
		h = int32(rl.GetScreenHeight())

		rl.UpdateCamera(&camera)

		rc := color(defaultColor)
		rl.ClearBackground(rc)

		{ // draw 3D model
			rl.BeginMode3D(camera)

			rl.DrawCube(cubePosition, 2.0, 2.0, 2.0, rl.Red)
			rl.DrawCubeWires(cubePosition, 2.0, 2.0, 2.0, rl.Maroon)
			rl.DrawCube(cubePosition, 9.0, 1.0, 1.0, rl.Red)
			rl.DrawCubeWires(cubePosition, 9.0, 1.0, 1.0, rl.Maroon)
			rl.DrawGrid(10, 1.0)

			rl.EndMode3D()
		}

		// draw gizmo
		DrawGizmo()

		{ // draw tui
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
func DrawGizmo() { // Vector3 position) {
	// NOTE: RGB = XYZ
	     float lenght = 1.0f;
	
	     rl.PushMatrix();
	         rl.Translatef(position.x, position.y, position.z);
	         //rlRotatef(rotation, 0, 1, 0);
	         rl.Scalef(lenght, lenght, lenght);
	
	         rl.Begin(rl.Lines);
	             rl.Color3f(1.0, 0.0, 0.0); rl.Vertex3f(0.0, 0.0, 0.0);
	             rl.Color3f(1.0, 0.0, 0.0); rl.Vertex3f(1.0, 0.0, 0.0);
	
	             rl.Color3f(0.0, 1.0, 0.0); rl.Vertex3f(0.0, 0.0, 0.0);
	             rl.Color3f(0.0, 1.0, 0.0); rl.Vertex3f(0.0, 1.0, 0.0);
	
	             rl.Color3f(0.0, 0.0, 1.0); rl.Vertex3f(0.0, 0.0, 0.0);
	             rl.Color3f(0.0, 0.0, 1.0); rl.Vertex3f(0.0, 0.0, 1.0);
	         rl.End();
	     rl.PopMatrix();
}

