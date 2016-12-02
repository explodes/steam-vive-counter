package main

const (
	categorySingleplayer      AppCategory = 2
	categoryMultiplayer                   = 1
	categoryOnlineMultiplayer             = 36
	categoryLocalMultiplayer              = 37
)

type NumberOfPlayers struct {
	Response struct {
		PlayerCount int `json:"player_count"`
		//Result      int `json:"result"`
	} `json:"response"`
}

type AppInfoById map[string]AppInfo

type AppCategory int

type AppInfo struct {
	Success bool `json:"success"`
	Data    struct {
		//Type                string `json:"type"`
		Name string `json:"name"`
		//SteamAppid          int `json:"steam_appid"`
		//RequiredAge         int `json:"required_age"`
		//IsFree              bool `json:"is_free"`
		//DetailedDescription string `json:"detailed_description"`
		//AboutTheGame        string `json:"about_the_game"`
		//ShortDescription    string `json:"short_description"`
		//SupportedLanguages  string `json:"supported_languages"`
		//HeaderImage         string `json:"header_image"`
		//Website             string `json:"website"`
		//PcRequirements      struct {
		//			    Minimum string `json:"minimum"`
		//		    } `json:"pc_requirements"`
		//MacRequirements     []interface{} `json:"mac_requirements"`
		//LinuxRequirements   []interface{} `json:"linux_requirements"`
		//Developers          []string `json:"developers"`
		//Publishers          []string `json:"publishers"`
		//PriceOverview       struct {
		//			    Currency        string `json:"currency"`
		//			    Initial         int `json:"initial"`
		//			    Final           int `json:"final"`
		//			    DiscountPercent int `json:"discount_percent"`
		//		    } `json:"price_overview"`
		//Packages            []int `json:"packages"`
		//PackageGroups       []struct {
		//	Name                    string `json:"name"`
		//	Title                   string `json:"title"`
		//	Description             string `json:"description"`
		//	SelectionText           string `json:"selection_text"`
		//	SaveText                string `json:"save_text"`
		//	DisplayType             int `json:"display_type"`
		//	IsRecurringSubscription string `json:"is_recurring_subscription"`
		//	Subs                    []struct {
		//		Packageid                int `json:"packageid"`
		//		PercentSavingsText       string `json:"percent_savings_text"`
		//		PercentSavings           int `json:"percent_savings"`
		//		OptionText               string `json:"option_text"`
		//		OptionDescription        string `json:"option_description"`
		//		CanGetFreeLicense        string `json:"can_get_free_license"`
		//		IsFreeLicense            bool `json:"is_free_license"`
		//		PriceInCentsWithDiscount int `json:"price_in_cents_with_discount"`
		//	} `json:"subs"`
		//} `json:"package_groups"`
		//Platforms           struct {
		//			    Windows bool `json:"windows"`
		//			    Mac     bool `json:"mac"`
		//			    Linux   bool `json:"linux"`
		//		    } `json:"platforms"`
		Categories []struct {
			ID          AppCategory `json:"id"`
			Description string      `json:"description"`
		} `json:"categories"`
		//Genres              []struct {
		//	ID          string `json:"id"`
		//	Description string `json:"description"`
		//} `json:"genres"`
		//Screenshots         []struct {
		//	ID            int `json:"id"`
		//	PathThumbnail string `json:"path_thumbnail"`
		//	PathFull      string `json:"path_full"`
		//} `json:"screenshots"`
		//Movies              []struct {
		//	ID        int `json:"id"`
		//	Name      string `json:"name"`
		//	Thumbnail string `json:"thumbnail"`
		//	Webm      struct {
		//			  Num480 string `json:"480"`
		//			  Max    string `json:"max"`
		//		  } `json:"webm"`
		//	Highlight bool `json:"highlight"`
		//} `json:"movies"`
		//Achievements        struct {
		//			    Total int `json:"total"`
		//		    } `json:"achievements"`
		//ReleaseDate         struct {
		//			    ComingSoon bool `json:"coming_soon"`
		//			    Date       string `json:"date"`
		//		    } `json:"release_date"`
		//SupportInfo         struct {
		//			    URL   string `json:"url"`
		//			    Email string `json:"email"`
		//		    } `json:"support_info"`
		//Background          string `json:"background"`
	} `json:"data"`
}

func (a *AppInfo) hasCategory(cat AppCategory) bool {
	for _, category := range a.Data.Categories {
		if category.ID == cat {
			return true
		}
	}
	return false
}

func (a *AppInfo) IsSingleplayer() bool {
	return a.hasCategory(categorySingleplayer)
}

func (a *AppInfo) IsMultiplayer() bool {
	return a.hasCategory(categoryMultiplayer)
}

func (a *AppInfo) IsOnlineMultiplayer() bool {
	return a.hasCategory(categoryOnlineMultiplayer)
}

func (a *AppInfo) IsLocalMultiplayer() bool {
	return a.hasCategory(categoryLocalMultiplayer)
}
