# Fiber JWT
Đây là ví dụ được cải tiến từ [gofiber/jwt](https://github.com/gofiber/jwt)

gofiber/jwt sử dụng package [golang-jwt](https://github.com/golang-jwt/jwt)

## Hướng dẫn chạy thử
Code mẫu chỉ có duy nhật một user là 
```json
{
  "user": "john",
  "pass": "doe"
}
```

Login using username and password to retrieve a token.
```
curl --data "user=john&pass=doe" http://localhost:3000/login
```

Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE0NjE5NTcxMzZ9.RB3arc4-OyzASAaUhC2W3ReWaXAt_z2Fd3BN4aWTgEY"
}
```

## Khởi tạo session redis
```go
var (
	CookieNameForSessionID = "testcookiesession"
	Sess = session.New(session.Config{
		KeyLookup: "cookie:"+CookieNameForSessionID,		
		Expiration: time.Hour * 8760,
		Storage: connectRedis(),
	})
)
```
```go
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
```

## Khi login lưu session
```go
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
```

## Cấu hình JWT middleware

```go
app.Use(middleware)
```
```go
func middleware(c *fiber.Ctx) error{
	//Đăng ký kiểu trả về, token là string nên để một chuỗi string bất kỳ vào
	Sess.RegisterType("t")
	session, _ := Sess.Get(c)

	if tokenString,ok := session.Get("authen").(string);!ok{
		return handle_when_invalid_token(c, errors.New("authenticate fail"))
	}else{
		//parse ra claim từ token lấy ra ở session cookie
		if claims, err := parseTokenClaims(tokenString);err == nil{
			if claims["name"] != "John Doe"{
				return handle_when_invalid_token(c, errors.New("authenticate fail"))
			}
		}
	}
	return c.Next()
}
```

Hàm parseTokenClaim
```go
func parseTokenClaims(tokenString string) (jwt.MapClaims, error){
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{},func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil || !token.Valid{
		return nil, err
	}

	claims := token.Claims.(jwt.MapClaims)
	return claims, nil
}
```