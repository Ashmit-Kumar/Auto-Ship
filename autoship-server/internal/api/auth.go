// internal/api/auth.go
package api

import (
	"context"
	// "fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/db"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/bson"
)

// Signup handler
func Signup(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Check if user already exists
	var existingUser models.User
	err := db.UserCollection.FindOne(context.TODO(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "User already exists"})
	}

	// Hash password before saving
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error hashing password"})
	}
	user.Password = string(hashedPassword)

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Insert new user into MongoDB
	_, err = db.UserCollection.InsertOne(context.TODO(), user)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating user"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	return c.JSON(fiber.Map{"message": "User registered successfully", "token": token})
}

// Login handler
func Login(c *fiber.Ctx) error {
	var loginDetails struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&loginDetails); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user models.User
	// Look up user by email
	err := db.UserCollection.FindOne(context.TODO(), bson.M{"email": loginDetails.Email}).Decode(&user)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	// Compare the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginDetails.Password)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Error generating token"})
	}

	return c.JSON(fiber.Map{"message": "Login successful", "token": token})
}
// GitHubLogin redirects the user to GitHub's OAuth consent page
func GitHubLogin(c *fiber.Ctx) error {
	url := utils.GetGitHubAuthURL()
	return c.Redirect(url, fiber.StatusTemporaryRedirect)
}
// GitHubCallback handles GitHub's redirect with the auth code
func GitHubCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Authorization code not provided"})
	}

	// Exchange the code for an access token
	accessToken, err := utils.ExchangeCodeForAccessToken(code)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to exchange code for token"})
	}

	// Get the GitHub user info
	userInfo, err := utils.GetGitHubUserInfo(accessToken)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user info from GitHub"})
	}

	// Extract email (or fallback)
	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		email = userInfo["login"].(string) + "@github.com"
	}

	// Check if user exists
	var existingUser models.User
	err = db.UserCollection.FindOne(context.TODO(), bson.M{"email": email}).Decode(&existingUser)
	if err != nil {
		// Create new user
		newUser := models.User{
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Password:  "", // GitHub login only
		}
		insertResult, err := db.UserCollection.InsertOne(context.TODO(), newUser)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
		}
		newUser.ID = insertResult.InsertedID.(primitive.ObjectID)
		existingUser = newUser
	}

	// Generate JWT
	token, err := utils.GenerateJWT(existingUser.ID.Hex(), existingUser.Email)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate JWT"})
	}

	// Optional: set as cookie instead of query param
	// c.Cookie(&fiber.Cookie{
	//     Name:     "auth_token",
	//     Value:    token,
	//     HTTPOnly: true,
	//     Secure:   false, // true in production
	//     Path:     "/",
	// })

	// Redirect to frontend dashboard with token
	redirectURL := "http://localhost:3000/dashboard?token=" + token
	return c.Redirect(redirectURL, fiber.StatusFound)
}
