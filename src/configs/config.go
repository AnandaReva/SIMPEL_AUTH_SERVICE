package configs

var otpExpireTime int16 = 180    //s
var resetPassExpTime int16 = 300 //s
var PBKDF2Iterations int = 15000
var clientURL string = "http://localhost:3000"

func GetOTPExpireTime() int16 {
	return otpExpireTime

}

func GetResetPassExpTime() int16 {
	return resetPassExpTime
}

func GetPBKDF2Iterations() int {
	return PBKDF2Iterations
}

func GetClientURL() string {
	return clientURL
}
