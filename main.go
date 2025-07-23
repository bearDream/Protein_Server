package main

import (
	"Protein_Server/logger"
	profasacontrollers "Protein_Server/profasa/controllers"
	proteinflowcontrollers "Protein_Server/proteinflow/controllers"
	psgocontrollers "Protein_Server/psgo/controllers"
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
	// 测试队列
	//testQueue()
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

	// 只对需要鉴权的接口加 JwtVerify
	auth := router.Group("/")
	auth.Use(services.JwtVerify)
	{
		auth.GET("/getUserInfo", profasacontrollers.GetUserInformation)
		auth.GET("/getotheruser", profasacontrollers.GetOtherUser)
		auth.POST("/blast", profasacontrollers.Blast)
		auth.POST("/superimpose", profasacontrollers.Superimpose)
		auth.POST("/single", profasacontrollers.Single)

		// 其他需要鉴权的接口...
		auth.GET("/getAllUserNotMe", profasacontrollers.GetAllUserNotMe)
		auth.GET("/share/show", profasacontrollers.ShowShare)
		auth.GET("/getBlastList", profasacontrollers.GetBlastList)
	}

	// Define Routers
	profasa := router.Group("/profasa")
	{
		account := profasa.Group("/account")
		{
			account.POST("/register", profasacontrollers.Register)
			account.POST("/login", profasacontrollers.LogIn)
			account.POST("/forgetpassword", profasacontrollers.ForgetPassword)
			// Verify the Token
			account.Use(services.JwtVerify)
			account.GET("/getuserinformation", profasacontrollers.GetUserInformation)
			account.GET("/getotheruser", profasacontrollers.GetOtherUser)
		}
		protein := profasa.Group("/protein")
		{
			// Verify the Token
			protein.Use(services.JwtVerify)
			protein.POST("/sequencesearch", profasacontrollers.SequenceSearch)
			protein.POST("/structureprediction", profasacontrollers.StructurePrediction)
			protein.POST("/parameterscalculation", profasacontrollers.ParametersCalculation)
			protein.POST("/superimpose", profasacontrollers.Superimpose)
			protein.POST("/blast", profasacontrollers.Blast)
			protein.GET("/task", profasacontrollers.TaskList)
			protein.GET("/task/:id", profasacontrollers.TaskDetails)
			protein.DELETE("/task/:id", profasacontrollers.DeleteTask)
			protein.POST("/sharetask/:id/:userid", profasacontrollers.ShareTask)
			protein.POST("/sharetask/agree/:id", profasacontrollers.AgreeShare)
			protein.POST("/sharetask/reject/:id", profasacontrollers.RejectShare)
			protein.GET("/notes/:id", profasacontrollers.ViewNotes)
			protein.POST("/notes/:id", profasacontrollers.UpdateNotes)
		}
		system := profasa.Group("/system")
		{
			system.POST("/uploadfile", profasacontrollers.UploadFile)
			system.GET("/queuestatus", profasacontrollers.GetQueueStatus)
		}
	}
	psgo := router.Group("/psgo")
	{
		search := psgo.Group("/search")
		{
			search.GET("/useparameters", psgocontrollers.SearchByParameters)
			search.GET("/usenaturelanguage", psgocontrollers.SearchByNatureLanguage)
		}
		protein := psgo.Group("/protein")
		{
			protein.GET("/information/:id", psgocontrollers.GetInformation)
			protein.GET("/maxsize", psgocontrollers.MaxSize)
		}
	}
	proteinflow := router.Group("/proteinflow")
	{
		calculation := proteinflow.Group("/calculation")
		{
			calculation.POST("/renderpdb2obj", proteinflowcontrollers.RenderPDB2OBJ)
		}
		admin := proteinflow.Group("/admin")
		{
			admin.GET("/folddurationlist", proteinflowcontrollers.FoldDurationList)
			admin.GET("/parameter/list", proteinflowcontrollers.ParameterList)
			admin.GET("/parameter/export", proteinflowcontrollers.ParameterExport)
		}
	}
	// Listen 10010 port
	// router.Run(":10010")
	logger.Info("服务器启动在端口 10010")
	router.Run(":10010")
}
