package main

import (
	profasacontrollers "Protein_Server/profasa/controllers"
	proteinflowcontrollers "Protein_Server/proteinflow/controllers"
	psgocontrollers "Protein_Server/psgo/controllers"
	"Protein_Server/services"

	_ "Protein_Server/database"

	"github.com/gin-gonic/gin"
)

func main() {
	services.BackendProcess()
	// Create a Gin Server
	router := gin.Default()
	// CORS Middleware
	router.Use(services.CORS())
	// Define Routers
	profasa := router.Group("/profasa")
	{
		account := profasa.Group("/account")
		{
			account.POST("/register", profasacontrollers.Register)
			account.POST("/login", profasacontrollers.LogIn)
			account.POST("/forgetpassword", profasacontrollers.ForgetPassword)
			account.POST("/sendcode", profasacontrollers.SendCode)
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
	// Listen 8080 port
	router.Run(":8080")
}
