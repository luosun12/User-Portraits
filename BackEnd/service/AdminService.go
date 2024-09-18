package service

//func AdminRegister(c *gin.Context) {
//	db, err := database.InitDB()
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"message": "数据库连接失败,请重试",
//		})
//		fmt.Printf("register err:%v\n", err)
//		return
//	}
//	var sql = Controllers.SqlController{DB: db}
//	var administrator etc.Admininfo
//	newName := c.PostForm("adminname")
//	newPswd := c.PostForm("password")
//	if newName == "" {
//		c.JSON(http.StatusBadRequest, gin.H{
//			"message": "用户名不能为空",
//		})
//		fmt.Printf("register err:bad newname\n")
//		return
//	}
//}
