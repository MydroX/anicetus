package middlewares

// func AuthMiddleware(cfg config.Config) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Extract token from header
// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid authorization header"})
// 			return
// 		}

// 		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

// 		claims, err := jwt.ParseAccessToken(tokenString)
// 		if err != nil {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			return
// 		}

// 		// check claims
// 		if claims.TokenType != jwt.AccessToken {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
// 			return
// 		}

// 		if claims.Exp < time.Now().Unix() {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
// 			return
// 		}

// 		c.Next()
// 	}
// }
