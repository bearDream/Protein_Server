package main

import (
	"Protein_Server/logger"
	profasacontrollers "Protein_Server/profasa/controllers"
	"Protein_Server/services"

	_ "Protein_Server/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志系统
	logger.Info("启动蛋白质服务器...")

	// 测试蛋白质参数计算正确性
	//if err := services.ReadAndCalcParameterExcel(); err != nil {
	//	logger.Error("批量计算Excel参数失败: %v", err)
	//}
	// 启动队列调度器
	queueScheduler := services.GetGlobalQueueScheduler()
	queueScheduler.Start()
	defer queueScheduler.Stop()

	//services.BackendProcess()
	// Create a Gin Server
	router := gin.Default()
	// CORS Middleware
	router.Use(services.CORS())

	// Static file services
	router.Static("/psgo/models", "../PROFASA-PDB-GO/data")
	router.Static("/psgo/imgs", "../PROFASA-PDB-GO/pdb_imgs")
	router.Static("/psgo/ramachandran", "../PROFASA-PDB-GO/ramachandran_plots")
	router.Static("/models", "static/models")
	router.Static("/imgs", "static/imgs")

	router.POST("/register", profasacontrollers.Register)
	router.POST("/logIn", profasacontrollers.LogIn)
	router.POST("/forgetpassword", profasacontrollers.ForgetPassword)
	router.POST("/uploadPDB", profasacontrollers.UploadPDB)
	router.POST("/uploadfasta", profasacontrollers.UploadFasta)
	router.POST("/pdb2x3d", profasacontrollers.PDB2X3D)
	router.POST("/pdb2obj", profasacontrollers.PDB2OBJ)
	router.POST("/pdb2fbx", profasacontrollers.PDB2FBX)
	// PSGO 和 admin 的相关接口
	router.GET("/searchPdbByParam", profasacontrollers.SearchPdbByParam)
	router.GET("/getPDBInformationById", profasacontrollers.GetPDBInformationById)
	router.GET("/getSeqTimeTable", profasacontrollers.GetSeqTimeTable)
	router.GET("/getPDBParameterList", profasacontrollers.GetPDBParameterList)
	router.GET("/calcAllPDBparams", profasacontrollers.CalcAllPDBParams)

	// 只对需要鉴权的接口加 JwtVerify
	auth := router.Group("/")
	auth.Use(services.JwtVerify)
	{
		auth.GET("/getUserInfo", profasacontrollers.GetUserInformation)
		auth.GET("/getotheruser", profasacontrollers.GetOtherUser)
		auth.POST("/blast", profasacontrollers.Blast)
		auth.POST("/fold", profasacontrollers.Fold)
		auth.POST("/superimpose", profasacontrollers.Superimpose)
		auth.POST("/single", profasacontrollers.Single)

		// 其他需要鉴权的接口...
		auth.GET("/getAllUserNotMe", profasacontrollers.GetAllUserNotMe)
		auth.GET("/share/show", profasacontrollers.ShowShare)
		auth.GET("/getBlastList", profasacontrollers.GetBlastList)
		auth.POST("/getBlastResult", profasacontrollers.GetBlastResult)
		auth.POST("/shareBlast", profasacontrollers.ShareBlast)
		auth.POST("/share/agree", profasacontrollers.AgreeShareBlast)
		auth.POST("/share/refuse", profasacontrollers.RefuseShareBlast)
		auth.POST("/viewNote", profasacontrollers.ViewNote)
		auth.POST("/getAllModelNotMe", profasacontrollers.GetAllModelNotMe)
		auth.POST("/updateNote", profasacontrollers.UpdateNote)
	}

	// Listen 10010 port
	// router.Run(":10010")
	logger.Info("服务器启动在端口 10010")
	router.Run(":10010")
}
