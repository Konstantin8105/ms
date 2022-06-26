package ms

import (
	_ "image/png"
	"log"
	"runtime"

	"github.com/go-gl/gl/all-core/gl"
	 // "github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/nullboundary/glfont"
)

const windowWidth = 600
const windowHeight = 400

func init() {
	runtime.LockOSThread()
}

// func text() {
//
// 	if err := glfw.Init(); err != nil {
// 		log.Fatalln("failed to initialize glfw:", err)
// 	}
// 	defer glfw.Terminate()
//
// 	glfw.WindowHint(glfw.Resizable, glfw.True)
// 	glfw.WindowHint(glfw.ContextVersionMajor, 3)
// 	glfw.WindowHint(glfw.ContextVersionMinor, 2)
// 	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
// 	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
//
// 	window, _ := glfw.CreateWindow(int(windowWidth), int(windowHeight), "glfontExample", glfw.GetPrimaryMonitor(), nil)
//
// 	window.MakeContextCurrent()
// 	glfw.SwapInterval(1)
//
// 	if err := gl.Init(); err != nil {
// 		panic(err)
// 	}
//
// 	//load font (fontfile, font scale, window width, window height
// 	font, err := glfont.LoadFont("Go-Bold.ttf", int32(10), windowWidth, windowHeight)
// 	if err != nil {
// 		log.Panicf("LoadFont: %v", err)
// 	}
//
// 	gl.Enable(gl.DEPTH_TEST)
// 	gl.DepthFunc(gl.LESS)
// 	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
//
// 	for !window.ShouldClose() {
// 		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
//
//      //set color and draw text
// 		font.SetColor(1.0, 1.0, 1.0, 1.0) //r,g,b,a font color
// 		font.Printf(80, 80, 1.0, "PI=3.14 Hello, 世界 Добро пожаловать в Википедию 0 1 2 3 4 5 6 7 8 9 A B C D E F  Lorem ipsum dolor sit amet, consectetur adipiscing elit.") //x,y,scale,string,printf args
//
// 		window.SwapBuffers()
// 		glfw.PollEvents()
//
// 	}
// }

const SampleString = "PI=3.1415926 Hello, 世界 Добро пожаловать в Википедию 0 1 2 3 4 5 6 7 8 9 A B C D E F"

// var fonts [16]*gltext.Font

// loadFont loads the specified font at the given scale.
// func loadFont(file string, scale int32) (*gltext.Font, error) {
// 	fd, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	defer fd.Close()
//
// 	return gltext.LoadTruetype(fd, scale, 32, 127, gltext.LeftToRight)
// }

// drawString draws the same string for each loaded font.
// func drawString(x, y float32, str string) error {
// 	for i := range fonts {
// 		if fonts[i] == nil {
// 			continue
// 		}
//
// 		// We need to offset each string by the height of the
// 		// font. To ensure they don't overlap each other.
// 		_, h := fonts[i].GlyphBounds()
// 		y := y + float32(i*h)
//
// 		// Draw a rectangular backdrop using the string's metrics.
// 		sw, sh := fonts[i].Metrics(SampleString)
// 		gl.Color4f(0.1, 0.1, 0.1, 0.7)
// 		gl.Rectf(x, y, x+float32(sw), y+float32(sh))
//
// 		// Render the string.
// 		gl.Color4f(1, 1, 1, 1)
// 		err := fonts[i].Printf(x, y, str)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }

// initGL initializes GLFW and OpenGL.
// func initGL() error {
// 	err := glfw.Init()
// 	if err != nil {
// 		return err
// 	}
//
// 	err = glfw.OpenWindow(640, 480, 8, 8, 8, 8, 0, 0, glfw.Windowed)
// 	if err != nil {
// 		glfw.Terminate()
// 		return err
// 	}
//
// 	glfw.SetWindowTitle("go-gl/gltext: Truetype font example")
// 	glfw.SetSwapInterval(1)
// 	glfw.SetWindowSizeCallback(onResize)
// 	glfw.SetKeyCallback(onKey)
//
// 	errno := gl.Init()
// 	if errno != gl.NO_ERROR {
// 		str, err := glu.ErrorString(errno)
// 		if err != nil {
// 			return fmt.Errorf("Unknown openGL error: %d", errno)
// 		}
// 		return fmt.Errorf(str)
// 	}
//
// 	gl.Disable(gl.DEPTH_TEST)
// 	gl.Disable(gl.LIGHTING)
// 	gl.ClearColor(0.2, 0.2, 0.23, 0.0)
// 	return nil
// }

// onKey handles key events.
// func onKey(key, state int) {
// 	if key == glfw.KeyEsc {
// 		glfw.CloseWindow()
// 	}
// }

// onResize handles window resize events.
// func onResize(w, h int) {
// 	if w < 1 {
// 		w = 1
// 	}
//
// 	if h < 1 {
// 		h = 1
// 	}
//
// 	gl.Viewport(0, 0, wid, h)
// 	gl.MatrixMode(gl.PROJECTION)
// 	gl.LoadIdentity()
// 	gl.Ortho(0, float64(w), float64(h), 0, 0, 1)
// 	gl.MatrixMode(gl.MODELVIEW)
// 	gl.LoadIdentity()
// }

var (
	rotationX float32
	rotationY float32
)

const width, height = 800, 600

// func init() {
// 	// GLFW event handling must run on the main OS thread
// 	runtime.LockOSThread()
// }

func M3() {
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	window, err := glfw.CreateWindow(width, height, "Cube", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	// file :=  "/home/konstantin/.fonts/Go-Bold.ttf"
	file :=  "/home/konstantin/.fonts/PTM55F.ttf"

	//load font (fontfile, font scale, window width, window height
	font, err := glfont.LoadFont(file, int32(20), windowWidth, windowHeight)
	if err != nil {
		panic(err)
		// log.Panicf("LoadFont: %v", err)
	}

	// Load the same font at different scale factors and directions.
	// 	for i := range fonts {
	// 		fonts[i], err = loadFont(file, int32(9+i))
	// 		if err != nil {
	// 			log.Printf("LoadFont: %v", err)
	// 			return
	// 		}
	//
	// 		defer fonts[i].Release()
	// 	}

	//setupScene()
	for !window.ShouldClose() {
		//drawScene()

		// err = drawString(10, 10, SampleString)
		// if err != nil {
		// 	log.Printf("Printf: %v", err)
		// 	return
		// }

		font.SetColor(1.0, 1.0, 1.0, 1.0) //r,g,b,a font color

		font.Printf(80, 80, 1.0, "PI=3.14 Hello, \n世界 Добро пожаловать в Википедию\n 0 1 2 3 4 5 6 7 8 9 A B C D E F\n  Lorem ipsum dolor sit amet, consectetur adipiscing elit.") //x,y,scale,string,printf args

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func setupScene() {
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.LIGHTING)

	gl.ClearColor(0.5, 0.5, 0.5, 0.0)
	gl.ClearDepth(1)
	gl.DepthFunc(gl.LEQUAL)

	ambient := []float32{0.5, 0.5, 0.5, 1}
	diffuse := []float32{1, 1, 1, 1}
	lightPosition := []float32{-5, 5, 10, 0}
	gl.Lightfv(gl.LIGHT0, gl.AMBIENT, &ambient[0])
	gl.Lightfv(gl.LIGHT0, gl.DIFFUSE, &diffuse[0])
	gl.Lightfv(gl.LIGHT0, gl.POSITION, &lightPosition[0])
	gl.Enable(gl.LIGHT0)

	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	f := float64(width)/height - 1
	gl.Frustum(-1-f, 1+f, -1, 1, 1.0, 10.0)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func drawScene() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Translatef(0, 0, -3.0)
	gl.Rotatef(rotationX, 1, 0, 0)
	gl.Rotatef(rotationY, 0, 1, 0)

	rotationX += 0.5
	rotationY += 0.5

	gl.Color4f(1, 1, 1, 1)

	gl.Begin(gl.QUADS)

	gl.Normal3f(0, 0, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, -1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, -1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, 1)

	gl.Normal3f(0, 0, -1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, 1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, 1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, -1)

	gl.Normal3f(0, 1, 0)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, 1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, -1)

	gl.Normal3f(0, -1, 0)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, -1, -1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, 1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, 1)

	gl.Normal3f(1, 0, 0)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(1, -1, -1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(1, 1, -1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(1, 1, 1)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(1, -1, 1)

	gl.Normal3f(-1, 0, 0)
	gl.TexCoord2f(0, 0)
	gl.Vertex3f(-1, -1, -1)
	gl.TexCoord2f(1, 0)
	gl.Vertex3f(-1, -1, 1)
	gl.TexCoord2f(1, 1)
	gl.Vertex3f(-1, 1, 1)
	gl.TexCoord2f(0, 1)
	gl.Vertex3f(-1, 1, -1)

	gl.End()
}
