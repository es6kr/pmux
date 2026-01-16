package server

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/net/websocket"

	"github.com/es6kr/pmux/internal/session"
	"github.com/es6kr/pmux/view"
)

// Start starts the web server
func Start(port int) error {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files
	e.Static("/static", "static")

	// Routes
	e.GET("/", handleIndex)
	e.GET("/sessions", handleSessions)
	e.POST("/sessions", handleCreateSession)
	e.DELETE("/sessions/:channel", handleDeleteSession)
	e.GET("/ws/:channel", handleWebSocket)

	// API routes
	api := e.Group("/api")
	api.GET("/sessions", handleAPIListSessions)
	api.POST("/sessions/:channel/keys", handleAPISendKeys)

	fmt.Printf("Starting pmux web UI at http://localhost:%d\n", port)
	return e.Start(fmt.Sprintf(":%d", port))
}

// Render helper for templ components
func render(c echo.Context, component templ.Component) error {
	return component.Render(c.Request().Context(), c.Response().Writer)
}

func handleIndex(c echo.Context) error {
	sessions, _ := session.List()
	return render(c, view.Index(sessions))
}

func handleSessions(c echo.Context) error {
	sessions, _ := session.List()
	return render(c, view.SessionList(sessions))
}

func handleCreateSession(c echo.Context) error {
	channel := c.FormValue("channel")
	if channel == "" {
		channel = "main"
	}

	_, err := session.Create(channel)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sessions, _ := session.List()
	return render(c, view.SessionList(sessions))
}

func handleDeleteSession(c echo.Context) error {
	channel := c.Param("channel")
	if err := session.Kill(channel); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sessions, _ := session.List()
	return render(c, view.SessionList(sessions))
}

func handleWebSocket(c echo.Context) error {
	channel := c.Param("channel")

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		// Simple echo for now - will be replaced with PTY
		for {
			var msg string
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				break
			}

			// Send keys to tmux session
			if err := session.SendKeys(channel, msg); err != nil {
				websocket.Message.Send(ws, fmt.Sprintf("Error: %v", err))
				continue
			}

			// Capture and send back pane content
			output, _ := session.CapturePane(channel)
			websocket.Message.Send(ws, output)
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func handleAPIListSessions(c echo.Context) error {
	sessions, err := session.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, sessions)
}

func handleAPISendKeys(c echo.Context) error {
	channel := c.Param("channel")

	var body struct {
		Keys string `json:"keys"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid body"})
	}

	if err := session.SendKeys(channel, body.Keys); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
