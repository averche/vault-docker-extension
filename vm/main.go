package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/hashicorp/vault-client-go"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
)

var client *vault.Client

func main() {
	var socketPath string
	flag.StringVar(&socketPath, "socket", "/run/guest/volumes-service.sock", "Unix domain socket to listen on")
	flag.Parse()

	os.RemoveAll(socketPath)

	c, err := vault.NewClient(vault.Configuration{
		BaseAddress: "http://vault:8201",
	})
	if err != nil {
		log.Fatal(err)
	}
	client = c
	client.SetToken("root")

	logrus.New().Infof("Starting listening on %s\n", socketPath)
	router := echo.New()
	router.HideBanner = true

	startURL := ""

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal(err)
	}
	router.Listener = ln

	router.GET("/get-secrets", getSecrets)
	router.GET("/get-secret", getSecret)
	router.POST("/write-secret", writeSecret)

	log.Fatal(router.Start(startURL))
}

func getSecrets(ctx echo.Context) error {
	/* */ log.Println("getSecrets: started")
	defer log.Println("getSecrets: done")

	resp, err := client.List(ctx.Request().Context(), "/secret/metadata")
	if err != nil {
		return fmt.Errorf("error sending request to vault :%w", err)
	}

	keys, ok1 := resp.Data["keys"]
	if !ok1 {
		return fmt.Errorf("no secrets found")
	}

	return ctx.JSON(http.StatusOK, keys)
}

func getSecret(ctx echo.Context) error {
	/* */ log.Println("getSecret:", ctx.QueryParam("key"), "started")
	defer log.Println("getSecret:", ctx.QueryParam("key"), "done")

	key := ctx.QueryParam("key")

	resp, err := client.Read(ctx.Request().Context(), fmt.Sprintf("/secret/data/%s", key))
	if err != nil {
		return fmt.Errorf("error sending request to vault :%w", err)
	}

	return ctx.JSON(http.StatusOK, resp.Data["data"])
}

func writeSecret(ctx echo.Context) error {
	/* */ log.Println("writeSecret:", ctx.QueryParam("key"), "started")
	defer log.Println("writeSecret:", ctx.QueryParam("key"), "done")

	defer ctx.Request().Body.Close()

	key := ctx.QueryParam("key")

	b, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return err
	}

	resp, err := client.WriteFromBytes(
		ctx.Request().Context(),
		fmt.Sprintf("/secret/data/%s", key),
		[]byte(fmt.Sprintf(`{"data": %s}`, string(b))),
	)
	if err != nil {
		return fmt.Errorf("error sending request to vault :%w", err)
	}

	return ctx.JSON(http.StatusOK, resp.Data["data"])
}
