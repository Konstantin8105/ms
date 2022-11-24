//go:build ignore

package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {

	// Initialization
	//--------------------------------------------------------------------------------------
	const (
		screenWidth  = 800
		screenHeight = 450
	)

	rl.InitWindow(screenWidth, screenHeight, "raylib [core] example - 3d camera free")

	// Define the camera to look into our 3d world
	var camera rl.Camera
	camera.Position = rl.Vector3{10.0, 10.0, 10.0}
	camera.Target = rl.Vector3{0.0, 0.0, 0.0}
	camera.Up = rl.Vector3{0.0, 1.0, 0.0}
	camera.Fovy = 45.0
	camera.Projection = rl.CameraPerspective

	var cubePosition rl.Vector3
	var cubeScreenPosition rl.Vector2

	rl.SetCameraMode(camera, rl.CameraFree) // Set a free camera mode

	rl.SetTargetFPS(60) // Set our game to run at 60 frames-per-second
	//--------------------------------------------------------------------------------------

	// Measure string size for Font
	// Vector2 MeasureTextEx(Font font, const char *text, float fontSize, float spacing);
	// func MeasureTextEx(font Font, text string, fontSize float32, spacing float32) Vector2 {
	font := rl.GetFontDefault()
	var fontSize float32 = 20.0
	var fontSpacing float32 = 1.0
	glyphSize := rl.MeasureTextEx(font, "*", fontSize, fontSpacing)
	glyphSize.X += 5
	glyphSize.Y += 5

	// Main game loop
	for !rl.WindowShouldClose() { // Detect window close button or ESC key
		// Update
		//----------------------------------------------------------------------------------
		rl.UpdateCamera(&camera)

		// Calculate cube screen space position (with a little offset to be in top)
		cubeScreenPosition = rl.GetWorldToScreen(
			rl.Vector3{
				cubePosition.X,
				cubePosition.Y + 2.5,
				cubePosition.Z,
			}, camera)
		//----------------------------------------------------------------------------------

		// Draw
		//----------------------------------------------------------------------------------
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.BeginMode3D(camera)

		rl.DrawCube(cubePosition, 2.0, 2.0, 2.0, rl.Red)
		rl.DrawCubeWires(cubePosition, 2.0, 2.0, 2.0, rl.Maroon)

		rl.DrawGrid(10, 1.0)

		rl.EndMode3D()

		rl.DrawText(
			"Enemy: 100 / 100",
			int32(cubeScreenPosition.X-float32(rl.MeasureText("Enemy: 100/100", 20)/2)),
			int32(cubeScreenPosition.Y),
			20,
			rl.Black,
		)
		rl.DrawText(
			"Text is always on top of the cube",
			(screenWidth-rl.MeasureText("Text is always on top of the cube", 20))/2,
			25,
			20,
			rl.Gray,
		)

		for r := 0; r < 10; r++ {
			for c := 0; c < 20; c++ {
				bg := rl.Yellow
				if (r+c)%3 == 0 {
					bg = rl.Gray
				}
				if (r+c)%2   == 1 {
					bg = rl.Blue
				}
				// Draw a color-filled rectangle
				// void DrawRectangle(int posX, int posY, int width, int height, Color color);
				rl.DrawRectangle(
					int32(glyphSize.X*float32(c)), // x
					int32(glyphSize.Y*float32(r)), // y
					int32(glyphSize.X),
					int32(glyphSize.Y),
					bg,
				)

				text := "T"
				if r%2 == 1 || c%2 == 0 {
					text = "p"
				}
				if r%2 == 1 && c%2 == 1 {
					text = "b"
				}

				// Draw text (using default font)
				// void DrawText(const char *text, int posX, int posY, int fontSize, Color color);
				rl.DrawText(
					text,
					int32(glyphSize.X*float32(c)), // x
					int32(glyphSize.Y*float32(r)), // y
					int32(fontSize),               // fontSize
					rl.Black,                       // color
				)
			}
		}

		rl.EndDrawing()
		//----------------------------------------------------------------------------------
	}

	// De-Initialization
	//--------------------------------------------------------------------------------------
	rl.CloseWindow() // Close window and OpenGL context
	//--------------------------------------------------------------------------------------
}
