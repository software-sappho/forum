package models

type PageData struct {
	IsLoggedIn       bool
	Username         string
	User             *User
	PostsCount       int
	CommentsCount    int
	LikesGiven       int
	DaysActive       int
	RecentPosts      []Post
	Post             Post
	Posts            []Post
	Comments         []Comment
	Categories       []Category
	SelectedCategory string
	FormData         map[string]string
	Errors           map[string]string
	Success          bool
	Message          string
	OfficeLocation   string
	BusinessHours    string
	CurrentPage      int
	TotalPages       int
	ErrorCode        int
	ErrorMessage     string
	LikedPosts       []Post
	LikedComments    []Comment
	NextURL          string
	PrevURL          string
}
