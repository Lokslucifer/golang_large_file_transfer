package main

import (
	"context"
	"large_fss/internals/constants"
	v1_handler "large_fss/internals/handlers/v1"
	middlewares "large_fss/internals/middleware"
	"large_fss/internals/repository"
	"large_fss/internals/services"
	"large_fss/internals/storage"
	// "path/filepath"

	"fmt"
	"log"

	// "time"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func ConnectDB() (*sqlx.DB, error) {

	dsn := constants.DBURL
	var err error
	DB, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}
	fmt.Println("✅ Connected to PostgreSQL!")

	return DB, nil
}

func ConnectAWSClient() *s3.Client {

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	fmt.Println("✅ Connected to AWS!")
	return s3Client
}
func main() {

	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Header("Expires", "0")
		c.Header("Pragma", "no-cache")
		c.Header("Surrogate-Control", "no-store")
		c.Next()
	})
	r.Static("/static", "./static")
	r.LoadHTMLGlob("./templates/*.html")
	//loading dot env
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		return
	}

	//connecting database
	Db, err := ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}
	defer Db.Close()
	postgres := repository.NewPostgresSQLDB(Db)

	filestorage := storage.NewLocalStorage("./Local_storage")

	jwtservice, err := services.NewJWTService()
	if err != nil {
		log.Fatalf("failed to create JWT service: %v", err)
	}
	
	mainservice := services.NewService(jwtservice, postgres, filestorage)
	go mainservice.CleanupService()

	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{})
	})
	r.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login_page.html", gin.H{})
	})
	r.GET("/share/:transferid", func(c *gin.Context) {
		transferID := c.Param("transferid")
		c.HTML(200, "share.html", gin.H{
			"TransferID": transferID,
		})
	})

	r.GET("/view/transfers/", middlewares.AuthorizationMiddleware(jwtservice), func(c *gin.Context) {

		c.HTML(200, "viewtransfers.html", gin.H{})
	})
	handler := v1_handler.NewHandler(mainservice)
	backend := r.Group("/api")
	backend.POST("/login", handler.LoginHandler)   //checked
	backend.POST("/signup", handler.SignupHandler) //checked
	publicTransferGroup := backend.Group("/transfer")
	publicTransferGroup.GET("/share/:transferid", handler.GetTransferInfoHandler)
	publicTransferGroup.GET("/download/file/:fileid", handler.FileDownloaderHandler)
	publicTransferGroup.GET("/download/transfer/:transferid", handler.TransferDownloaderHandler)

	protected := backend.Group("/auth") //checked
	protected.Use(middlewares.AuthorizationMiddleware(mainservice.JwtService))

	{
		protectedTransferRoutes := protected.Group("/transfer")
		protectedTransferRoutes.POST("/new", handler.CreateTransferHandler)
		protectedTransferRoutes.POST("/upload", handler.UploadChunkHandler)
		protectedTransferRoutes.POST("/cancel", handler.CancelTransferHandler)

		protectedTransferRoutes.POST("/assemble", handler.AssembleFileHandler)
		protectedTransferRoutes.GET("/successchunk/:transferid", handler.GetAllUploadedChunksIndexHandler)

		protectedTransferRoutes.DELETE("/delete/:transferid", handler.DeleteTransferHandler)
		protectedTransferRoutes.GET("/all", handler.GetAllTransfersHandler)
		protectedTransferRoutes.PUT("/update", handler.UpdateTransferHandler)

	}

	r.Run(constants.DefaultPort) // http://localhost:8081

}
