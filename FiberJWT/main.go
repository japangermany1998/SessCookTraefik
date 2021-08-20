package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/storage/redis"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt"
)

var SecretKey = []byte("TuyệtMật")
var (
	CookieNameForSessionID = "testcookiesession"
	Sess = session.New(session.Config{
		KeyLookup: "cookie:"+CookieNameForSessionID,		
		Expiration: time.Hour * 8760,
		Storage: connectRedis(),
	})
)

func main() {
	app := fiber.New()
	// Login route
	app.Post("/login", login)
	// Unauthenticated route
	app.Get("/", accessible)

	// Middleware sẽ lấy token từ session parse ra claims và xác thực
	app.Use(middleware)

	//Authenticate JWT sử dụng cho Traefik ForwardAuth
	app.Get("/auth", authenticate)

	app.Listen(":3000")
}

func connectRedis() *redis.Storage{
// Hiện đóng gói docker chưa kết nối được
	store := redis.New(redis.Config{
		Host:     "localhost",
		Port:     6379,
		Username: "",
		Password: "123",
		Database: 1,
		Reset:    false,
	})
	return store
}

func middleware(c *fiber.Ctx) error{
	//Đăng ký kiểu trả về, token là string nên để một chuỗi string bất kỳ vào
	Sess.RegisterType("t")
	session, _ := Sess.Get(c)
	log.Print(session.Get("authen"))

	if tokenString,ok := session.Get("authen").(string);!ok{
		log.Println(ok)

		return handle_when_invalid_token(c, errors.New("authenticate fail"))
	}else{
		log.Println(ok)
		//parse ra claim từ token lấy ra ở session cookie
		if claims, err := parseTokenClaims(tokenString);err == nil{
			if claims["name"] != "John Doe"{
				return handle_when_invalid_token(c, errors.New("authenticate fail"))
			}
		}
	}
	return c.Next()
}

func login(c *fiber.Ctx) error {
	user := c.FormValue("user")
	pass := c.FormValue("pass")
	// Throws Unauthorized error
	if user != "john" || pass != "doe" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "John Doe"
	claims["role"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString(SecretKey)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	//Lưu session
	session, _ := Sess.Get(c)
	session.Set("authen", t)
	session.Save()
	
	return c.JSON(fiber.Map{"token": t})
}

func accessible(c *fiber.Ctx) error {
	return c.SendString("Accessible")
}

func authenticate(c *fiber.Ctx) error {
	fmt.Println("***authenticate")
	fmt.Println(c)
	fmt.Println("BaseURL", c.BaseURL())
	fmt.Println("c.Request.URI", c.Request().URI())
	return c.Status(200).SendString("Authenticated!")
}
