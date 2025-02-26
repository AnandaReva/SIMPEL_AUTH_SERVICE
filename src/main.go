package main

import (
	//"auth_service/configs"

	"auth_service/db"
	"auth_service/handlers"
	"auth_service/logger"

	//	"auth_service/mail"
	"auth_service/middlewares"
	"auth_service/rds"

	//"fmt"

	"net/http"
	"os"
	"strconv"
)

// func generateReferenceID(timer int64) string {

// 	timeBase36 := strconv.FormatUint(uint64(timer), 36)
// 	randString, err := utils.RandomStringGenerator(8)
// 	if err != nil {
// 		randString = "12345678"
// 	}

// 	reference_id := timeBase36 + "." + randString // concate

// 	return reference_id

// }

func main() {
	// ENDPOINTS
	paths := make(map[string]func(http.ResponseWriter, *http.Request))

	//initialize database connection

	DBDRIVER := os.Getenv("DBDRIVER")
	DBNAME := os.Getenv("DBNAME")
	DBHOST := os.Getenv("DBHOST")
	DBUSER := os.Getenv("DBUSER")
	DBPASS := os.Getenv("DBPASS")
	DBPORT, err := strconv.Atoi(os.Getenv("DBPORT"))
	if err != nil {
		logger.Error("MAIN", "Failed to parse DBPORT, using default (5432), reason: ", err)
		DBPORT = 5432 // Default to 5432 if parsing fails
	}

	DBPOOLSIZE, err := strconv.Atoi(os.Getenv("DBPOOLSIZE"))
	if err != nil {
		logger.Warning("MAIN", "Failed to parse DBPOOLSIZE, using default (20)", err)
		DBPOOLSIZE = 20 // Default to 20 if parsing fails
	}

	if len(DBDRIVER) == 0 {
		logger.Error("DBDRIVER environment variable is required")
	}

	if len(DBNAME) == 0 {
		logger.Error("DBNAME environment variable is required")
	}

	if len(DBHOST) == 0 {
		logger.Error("DBHOST environment variable is required")
	}

	if len(DBUSER) == 0 {
		logger.Error("DBUSER environment variable is required")
	}

	if len(DBPASS) == 0 {
		logger.Error("DBPASS environment variable is required")
	}

	logger.Info("MAIN", "-----------POSTGRESQL CONF : ")
	logger.Info("MAIN", "DBDRIVER : ", DBDRIVER)
	logger.Info("MAIN", "DBHOST : ", DBHOST)
	logger.Info("MAIN", "DBPORT : ", DBPORT)
	logger.Debug("MAIN", "DBUSER : ", DBUSER)
	logger.Debug("MAIN", "DBPASS : ", DBPASS)
	logger.Info("MAIN", "DBNAME : ", DBNAME)
	logger.Info("MAIN", "DBPOOLSIZE : ", DBPOOLSIZE)

	err = db.InitDB(DBDRIVER, DBHOST, DBPORT, DBUSER, DBPASS, DBNAME, DBPOOLSIZE)
	if err != nil {
		logger.Error("MAIN", "ERROR !!! FAILED TO INITIATE DB POOL..", err)
		os.Exit(1)
	} else {
		logger.Info("MAIN", "Database Connection Pool Initated.")
	}

	logger.Info("MAIN", "-----------REDIS CONF : ")
	// log redis conf
	RDHOST := os.Getenv("RDHOST")
	RDPASS := os.Getenv("RDPASS")
	RDDB, errConv := strconv.Atoi(os.Getenv("RDDB"))

	if len(RDHOST) == 0 {
		logger.Error("RDHOST environment variable is required")
	}

	if len(RDPASS) == 0 {
		logger.Warning("RDPASS environment variable is required")
	}

	if errConv != nil {
		logger.Warning("MAIN", "Failed to parse RDDB, using default (0), reason: ", errConv)
		RDDB = 0 // Default to 0 if parsing fails
	}

	logger.Info("MAIN", "RDHOST : ", RDHOST)
	logger.Info("MAIN", "RDPASS : ", RDPASS)
	logger.Info("MAIN", "RDDB : ", RDDB)

	///////////////////////////////// POSTGRESQL ///////////////////////////////
	err = db.InitDB(DBDRIVER, DBHOST, DBPORT, DBUSER, DBPASS, DBNAME, DBPOOLSIZE)
	if err != nil {
		logger.Error("MAIN", "ERROR !!! FAILED TO INITIATE DB POOL..", err)
		os.Exit(1)
	} else {
		logger.Info("MAIN", "Database Connection Pool Initated.")
	}

	///////////////////////////////// REDIS ///////////////////////////////
	// Inisialisasi Redis hanya di main

	if err := rds.InitRedisConn(RDHOST, RDPASS, RDDB); err != nil {
		logger.Error("MAIN", "ERROR - Redis connection failed:", err)
		os.Exit(1)
	}

	///////////////////////////////// SMTP ///////////////////////////////
	logger.Info("MAIN", "-----------SMTP CONF : ")

	SMTPSERVER := os.Getenv("SMTPSERVER")
	SMTPPORT := os.Getenv("SMTPPORT")
	SMTPUSER := os.Getenv("SMTPUSER")
	SMTPPASS := os.Getenv("SMTPPASS")
	SMTPFROM := os.Getenv("SMTPFROM")

	logger.Info("MAIN", "SMTPSERVER : ", SMTPSERVER)
	logger.Info("MAIN", "SMTPPORT : ", SMTPPORT)
	logger.Info("MAIN", "SMTPUSER : ", SMTPUSER)
	logger.Info("MAIN", "SMTPPASS : ", SMTPPASS)
	logger.Info("MAIN", "SMTPFROM : ", SMTPFROM)

	if len(SMTPSERVER) == 0 {
		logger.Error("SMTPSERVER environment variable is required")
	}

	if len(SMTPPORT) == 0 {
		logger.Error("SMTPPORT environment variable is required")
	}

	if len(SMTPUSER) == 0 {
		logger.Error("SMPTPUSER environment variable is required")
	}

	if len(SMTPPASS) == 0 {
		logger.Error("SMPTPPASS environment variable is required")
	}

	if len(SMTPFROM) == 0 {
		logger.Error("SMTPFROM environment variable is required")
	}

	// Ambil email dari environment
	testEmail := SMTPUSER

	if testEmail == "" {
		logger.Error("MAIN", "SMTPUSER is not set in environment variables")
		os.Exit(1)
	}

	// Uji kirim email ke diri sendiri

	// testMessage := fmt.Sprintf("This is a test SMTP email \n service: %s  \n version: %s", configs.GetAppName(), configs.GetVersion())
	// err = mail.SendEmail(testEmail, "Test Email", testMessage)
	// if err != nil {
	// 	logger.Error("MAIN", "ERROR - Failed to send test email:", err)
	// 	os.Exit(1)
	// }

	// if err != nil {

	// 	logger.Error("MAIN", "ERROR - Failed to send email:", err)
	// 	os.Exit(1)

	// }

	paths["/"] = handlers.Greeting
	// send requestID and db conn as parameter
	paths["/login"] = handlers.Login
	paths["/register"] = handlers.Register
	paths["/logout"] = handlers.Logout
	paths["/verify-token"] = handlers.Verify_Token
	paths["/register/verify-otp"] = handlers.Register_Verify_OTP
	paths["/reset-password"] = handlers.Reset_Password
	paths["/reset-password/verify-url"] = handlers.Reset_Password_Verify_URL

	// Register endpoints with a multiplexer
	mux := http.NewServeMux()
	for path, handler := range paths {
		mux.HandleFunc(path, handler)
	}

	// Start server
	port := ":5000"
	logger.Info("INFO", "Starting server on http://localhost", port)
	if err := http.ListenAndServe(port, middlewares.CorsMiddleware(mux)); err != nil {
		logger.Error("MAIN", "Server failed: ", err)
	}
}
