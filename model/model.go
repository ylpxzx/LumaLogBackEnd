package model

type User struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	DisplayName        string `json:"display_name"`
	ThemePreference    string `json:"theme_preference"`
	LanguagePreference string `json:"language_preference"`
	DashboardViewMode  string `json:"dashboard_view_mode"`
	ShowTodayStatus    bool   `json:"show_today_status"`
	ShowCurrentStreak  bool   `json:"show_current_streak"`
	ShowLongestStreak  bool   `json:"show_longest_streak"`
	ShowCompletionRate bool   `json:"show_completion_rate"`
	ShowTotalCheckins  bool   `json:"show_total_checkins"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type Category struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	ColorTheme string `json:"color_theme"`
	SortOrder  int    `json:"sort_order"`
	IsDefault  bool   `json:"is_default"`
	IsHidden   bool   `json:"is_hidden"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type Item struct {
	ID                 int64  `json:"id"`
	UserID             int64  `json:"user_id"`
	CategoryID         int64  `json:"category_id"`
	CategoryName       string `json:"category_name"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	ColorTheme         string `json:"color_theme"`
	StartDate          string `json:"start_date"`
	EndDate            string `json:"end_date"`
	IsUnlimited        bool   `json:"is_unlimited"`
	DailyTargetCount   int    `json:"daily_target_count"`
	TimeMode           string `json:"time_mode"`
	ValidStartTime     string `json:"valid_start_time"`
	ValidEndTime       string `json:"valid_end_time"`
	AllowMakeup        bool   `json:"allow_makeup"`
	MakeupLimitDays    int    `json:"makeup_limit_days"`
	AllowExtraCheckins bool   `json:"allow_extra_checkins"`
	ShowOnDashboard    bool   `json:"show_on_dashboard"`
	SortOrder          int    `json:"sort_order"`
	ArchivedAt         string `json:"archived_at"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

type Checkin struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	ItemID      int64  `json:"item_id"`
	CheckinDate string `json:"checkin_date"`
	CheckinTime string `json:"checkin_time"`
	Count       int    `json:"count"`
	Note        string `json:"note"`
	Source      string `json:"source"`
	CreatedAt   string `json:"created_at"`
}

type HeatmapDay struct {
	Date      string `json:"date"`
	Count     int    `json:"count"`
	Level     int    `json:"level"`
	Completed bool   `json:"completed"`
}

type ItemStats struct {
	CurrentStreak  int     `json:"current_streak"`
	LongestStreak  int     `json:"longest_streak"`
	TotalCheckins  int     `json:"total_checkins"`
	CompletedDays  int     `json:"completed_days"`
	ExpectedDays   int     `json:"expected_days"`
	CompletionRate float64 `json:"completion_rate"`
}

type DashboardItem struct {
	Item       Item         `json:"item"`
	Stats      ItemStats    `json:"stats"`
	Heatmap    []HeatmapDay `json:"heatmap"`
	TodayCount int          `json:"today_count"`
	Status     string       `json:"status"`
}

type DashboardResponse struct {
	User       User            `json:"user"`
	Categories []Category      `json:"categories"`
	Items      []DashboardItem `json:"items"`
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token      string     `json:"token"`
	User       User       `json:"user"`
	Categories []Category `json:"categories,omitempty"`
}

type CategoryRequest struct {
	Name       *string `json:"name"`
	ColorTheme *string `json:"color_theme"`
	SortOrder  *int    `json:"sort_order"`
	IsHidden   *bool   `json:"is_hidden"`
}

type ItemRequest struct {
	CategoryID         *int64  `json:"category_id"`
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	ColorTheme         *string `json:"color_theme"`
	StartDate          *string `json:"start_date"`
	EndDate            *string `json:"end_date"`
	IsUnlimited        *bool   `json:"is_unlimited"`
	DailyTargetCount   *int    `json:"daily_target_count"`
	TimeMode           *string `json:"time_mode"`
	ValidStartTime     *string `json:"valid_start_time"`
	ValidEndTime       *string `json:"valid_end_time"`
	AllowMakeup        *bool   `json:"allow_makeup"`
	MakeupLimitDays    *int    `json:"makeup_limit_days"`
	AllowExtraCheckins *bool   `json:"allow_extra_checkins"`
	ShowOnDashboard    *bool   `json:"show_on_dashboard"`
	SortOrder          *int    `json:"sort_order"`
}

type CheckinRequest struct {
	Count  *int    `json:"count"`
	Note   *string `json:"note"`
	Source *string `json:"source"`
}

type PreferenceRequest struct {
	ThemePreference    *string `json:"theme_preference"`
	LanguagePreference *string `json:"language_preference"`
	DashboardViewMode  *string `json:"dashboard_view_mode"`
	ShowTodayStatus    *bool   `json:"show_today_status"`
	ShowCurrentStreak  *bool   `json:"show_current_streak"`
	ShowLongestStreak  *bool   `json:"show_longest_streak"`
	ShowCompletionRate *bool   `json:"show_completion_rate"`
	ShowTotalCheckins  *bool   `json:"show_total_checkins"`
}

type TokenPayload struct {
	UserID int64 `json:"user_id"`
	Exp    int64 `json:"exp"`
}

var DefaultCategories = []Category{
	{Name: "戒断", Slug: "quit", ColorTheme: "red", SortOrder: 10, IsDefault: true},
	{Name: "健康", Slug: "health", ColorTheme: "green", SortOrder: 20, IsDefault: true},
	{Name: "健身", Slug: "fitness", ColorTheme: "orange", SortOrder: 30, IsDefault: true},
	{Name: "学习", Slug: "study", ColorTheme: "blue", SortOrder: 40, IsDefault: true},
	{Name: "阅读", Slug: "reading", ColorTheme: "teal", SortOrder: 50, IsDefault: true},
	{Name: "工作", Slug: "work", ColorTheme: "gray", SortOrder: 60, IsDefault: true},
	{Name: "创作", Slug: "creative", ColorTheme: "purple", SortOrder: 70, IsDefault: true},
	{Name: "生活", Slug: "life", ColorTheme: "pink", SortOrder: 80, IsDefault: true},
}
