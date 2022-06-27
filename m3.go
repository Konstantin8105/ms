package ms

import (
	_ "image/png"
	"log"
	"runtime"

	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

func M3() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(1280, 780, "3D model", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1) // Enable vsync

	if err := gl.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)

	// ???

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.TEXTURE_2D)

	for !window.ShouldClose() {
		glfw.PollEvents()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.Begin(gl.TRIANGLES)
		gl.TexCoord2f(0.5, 1.0)
		gl.Vertex2f(-3, 3)
		gl.TexCoord2f(0.0, 0.0)
		gl.Vertex2f(-3, 0)
		gl.TexCoord2f(1.0, 0.0)
		gl.Vertex2f(0, 0)
		gl.End()
		gl.Flush()

		gl.ClearColor(0.5, 0.5, 0.5, 0.0)
		gl.ClearDepth(1)
		gl.DepthFunc(gl.LEQUAL)

		window.MakeContextCurrent()
		window.SwapBuffers()
	}
}
