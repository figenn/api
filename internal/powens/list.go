package powens

import "regexp"

type SubscriptionList struct {
	Name     string
	Category string
	Regex    *regexp.Regexp // Ajout d'un champ pour les expressions régulières
}

var SubscriptionNames = []SubscriptionList{
	// Streaming Vidéo
	{"PAYPAL", "Streaming Vidéo", regexp.MustCompile(`(?i)PAYPAL`)},
	{"Itunes", "Streaming Vidéo", regexp.MustCompile(`(?i)Itunes`)},
	{"Apple", "Streaming Vidéo", regexp.MustCompile(`(?i)NETFLIX`)},
	{"Netflix", "Streaming Vidéo", regexp.MustCompile(`(?i)Netflix`)},
	{"Disney+", "Streaming Vidéo", regexp.MustCompile(`(?i)Disney\+`)},
	{"Amazon Prime Video", "Streaming Vidéo", regexp.MustCompile(`(?i)Amazon.*Prime.*Video`)},
	{"Apple TV+", "Streaming Vidéo", regexp.MustCompile(`(?i)Apple.*TV\+`)},
	{"HBO Max", "Streaming Vidéo", regexp.MustCompile(`(?i)HBO.*Max`)},
	{"Hulu", "Streaming Vidéo", regexp.MustCompile(`(?i)Hulu`)},
	{"Paramount+", "Streaming Vidéo", regexp.MustCompile(`(?i)Paramount\+`)},
	{"Canal+", "Streaming Vidéo", regexp.MustCompile(`(?i)Canal\+`)},
	{"Crunchyroll", "Streaming Vidéo", regexp.MustCompile(`(?i)Crunchyroll`)},
	{"Funimation", "Streaming Vidéo", regexp.MustCompile(`(?i)Funimation`)},
	{"WOW Presents Plus", "Streaming Vidéo", regexp.MustCompile(`(?i)WOW.*Presents.*Plus`)},
	{"Mubi", "Streaming Vidéo", regexp.MustCompile(`(?i)Mubi`)},
	{"Peacock", "Streaming Vidéo", regexp.MustCompile(`(?i)Peacock`)},
	{"Rakuten TV", "Streaming Vidéo", regexp.MustCompile(`(?i)Rakuten.*TV`)},
	{"Molotov", "Streaming Vidéo", regexp.MustCompile(`(?i)Molotov`)},
	{"Starzplay", "Streaming Vidéo", regexp.MustCompile(`(?i)Starzplay`)},
	{"Shudder", "Streaming Vidéo", regexp.MustCompile(`(?i)Shudder`)},
	{"Filmotv", "Streaming Vidéo", regexp.MustCompile(`(?i)Filmotv`)},
	{"Ovid.tv", "Streaming Vidéo", regexp.MustCompile(`(?i)Ovid\.tv`)},
	{"Apple One", "Streaming Vidéo", regexp.MustCompile(`(?i)Apple.*One`)},

	// Streaming Musique
	{"Spotify", "Streaming Musique", regexp.MustCompile(`(?i)Spotify`)},
	{"Apple Music", "Streaming Musique", regexp.MustCompile(`(?i)Apple.*Music`)},
	{"Deezer", "Streaming Musique", regexp.MustCompile(`(?i)Deezer`)},
	{"YouTube Music", "Streaming Musique", regexp.MustCompile(`(?i)YouTube.*Music`)},
	{"Tidal", "Streaming Musique", regexp.MustCompile(`(?i)Tidal`)},
	{"Amazon Music", "Streaming Musique", regexp.MustCompile(`(?i)Amazon.*Music`)},
	{"Qobuz", "Streaming Musique", regexp.MustCompile(`(?i)Qobuz`)},
	{"SoundCloud Go+", "Streaming Musique", regexp.MustCompile(`(?i)SoundCloud.*Go\+`)},
	{"Napster", "Streaming Musique", regexp.MustCompile(`(?i)Napster`)},
	{"Boomplay", "Streaming Musique", regexp.MustCompile(`(?i)Boomplay`)},
	{"Anghami", "Streaming Musique", regexp.MustCompile(`(?i)Anghami`)},

	// Livres Audio & Presse
	{"Audible", "Livres Audio", regexp.MustCompile(`(?i)Audible`)},
	{"Scribd", "Livres Audio", regexp.MustCompile(`(?i)Scribd`)},
	{"Kindle Unlimited", "Livres Audio", regexp.MustCompile(`(?i)Kindle.*Unlimited`)},
	{"YouScribe", "Livres Audio", regexp.MustCompile(`(?i)YouScribe`)},
	{"PressReader", "Actualités & Presse", regexp.MustCompile(`(?i)PressReader`)},
	{"Spotify Audiobooks", "Livres Audio", regexp.MustCompile(`(?i)Spotify.*Audiobooks`)},
	{"Le Monde", "Actualités & Presse", regexp.MustCompile(`(?i)Le.*Monde`)},
	{"Le Figaro", "Actualités & Presse", regexp.MustCompile(`(?i)Le.*Figaro`)},
	{"Mediapart", "Actualités & Presse", regexp.MustCompile(`(?i)Mediapart`)},
	{"The New York Times", "Actualités & Presse", regexp.MustCompile(`(?i)The.*New.*York.*Times`)},
	{"Washington Post", "Actualités & Presse", regexp.MustCompile(`(?i)Washington.*Post`)},
	{"Les Échos", "Actualités & Presse", regexp.MustCompile(`(?i)Les.*Échos`)},

	// Stockage Cloud
	{"Google Drive", "Stockage Cloud", regexp.MustCompile(`(?i)Google.*Drive`)},
	{"Dropbox", "Stockage Cloud", regexp.MustCompile(`(?i)Dropbox`)},
	{"iCloud", "Stockage Cloud", regexp.MustCompile(`(?i)iCloud`)},
	{"OneDrive", "Stockage Cloud", regexp.MustCompile(`(?i)OneDrive`)},
	{"pCloud", "Stockage Cloud", regexp.MustCompile(`(?i)pCloud`)},
	{"Mega", "Stockage Cloud", regexp.MustCompile(`(?i)Mega`)},
	{"Amazon S3", "Stockage Cloud", regexp.MustCompile(`(?i)Amazon.*S3`)},
	{"Backblaze", "Stockage Cloud", regexp.MustCompile(`(?i)Backblaze`)},
	{"Sync.com", "Stockage Cloud", regexp.MustCompile(`(?i)Sync\.com`)},
	{"Box", "Stockage Cloud", regexp.MustCompile(`(?i)Box`)},
	{"iDrive", "Stockage Cloud", regexp.MustCompile(`(?i)iDrive`)},

	// Jeux Vidéo
	{"Xbox Game Pass", "Jeux Vidéo", regexp.MustCompile(`(?i)Xbox.*Game.*Pass`)},
	{"PlayStation Plus", "Jeux Vidéo", regexp.MustCompile(`(?i)PlayStation.*Plus`)},
	{"Nintendo Switch Online", "Jeux Vidéo", regexp.MustCompile(`(?i)Nintendo.*Switch.*Online`)},
	{"EA Play", "Jeux Vidéo", regexp.MustCompile(`(?i)EA.*Play`)},
	{"Ubisoft+", "Jeux Vidéo", regexp.MustCompile(`(?i)Ubisoft\+`)},
	{"GeForce Now", "Jeux Vidéo", regexp.MustCompile(`(?i)GeForce.*Now`)},
	{"Apple Arcade", "Jeux Vidéo", regexp.MustCompile(`(?i)Apple.*Arcade`)},
	{"Shadow", "Jeux Vidéo", regexp.MustCompile(`(?i)Shadow`)},
	{"Blacknut", "Jeux Vidéo", regexp.MustCompile(`(?i)Blacknut`)},
	{"Roblox Premium", "Jeux Vidéo", regexp.MustCompile(`(?i)Roblox.*Premium`)},
	{"Minecraft Realms", "Jeux Vidéo", regexp.MustCompile(`(?i)Minecraft.*Realms`)},
	{"Luna (Amazon Gaming)", "Jeux Vidéo", regexp.MustCompile(`(?i)Luna.*\(Amazon.*Gaming\)`)},
	{"Battle.net Pass", "Jeux Vidéo", regexp.MustCompile(`(?i)Battle\.net.*Pass`)},
	{"Final Fantasy XIV Subscription", "Jeux Vidéo", regexp.MustCompile(`(?i)Final.*Fantasy.*XIV.*Subscription`)},
	{"RuneScape Membership", "Jeux Vidéo", regexp.MustCompile(`(?i)RuneScape.*Membership`)},
	{"Dofus Premium", "Jeux Vidéo", regexp.MustCompile(`(?i)Dofus.*Premium`)},

	// Productivité
	{"Microsoft 365", "Productivité", regexp.MustCompile(`(?i)Microsoft.*365`)},
	{"Google Workspace", "Productivité", regexp.MustCompile(`(?i)Google.*Workspace`)},
	{"Notion", "Productivité", regexp.MustCompile(`(?i)Notion`)},
	{"Evernote", "Productivité", regexp.MustCompile(`(?i)Evernote`)},
	{"Trello", "Productivité", regexp.MustCompile(`(?i)Trello`)},
	{"Slack", "Productivité", regexp.MustCompile(`(?i)Slack`)},
	{"Zoom", "Productivité", regexp.MustCompile(`(?i)Zoom`)},
	{"Adobe Creative Cloud", "Productivité", regexp.MustCompile(`(?i)Adobe.*Creative.*Cloud`)},
	{"Grammarly", "Productivité", regexp.MustCompile(`(?i)Grammarly`)},
	{"Canva", "Productivité", regexp.MustCompile(`(?i)Canva`)},
	{"Monday.com", "Productivité", regexp.MustCompile(`(?i)Monday\.com`)},
	{"Asana", "Productivité", regexp.MustCompile(`(?i)Asana`)},
	{"ClickUp", "Productivité", regexp.MustCompile(`(?i)ClickUp`)},
	{"Todoist", "Productivité", regexp.MustCompile(`(?i)Todoist`)},
	{"Obsidian Sync", "Productivité", regexp.MustCompile(`(?i)Obsidian.*Sync`)},
	{"Basecamp", "Productivité", regexp.MustCompile(`(?i)Basecamp`)},
	{"Roam Research", "Productivité", regexp.MustCompile(`(?i)Roam.*Research`)},
	{"Bear Pro", "Productivité", regexp.MustCompile(`(?i)Bear.*Pro`)},
	{"Superhuman", "Productivité", regexp.MustCompile(`(?i)Superhuman`)},
	{"Dropbox Paper", "Productivité", regexp.MustCompile(`(?i)Dropbox.*Paper`)},

	// Fitness & Bien-être
	{"MyFitnessPal", "Fitness & Bien-être", regexp.MustCompile(`(?i)MyFitnessPal`)},
	{"Strava", "Fitness & Bien-être", regexp.MustCompile(`(?i)Strava`)},
	{"Fitbit Premium", "Fitness & Bien-être", regexp.MustCompile(`(?i)Fitbit.*Premium`)},
	{"Apple Fitness+", "Fitness & Bien-être", regexp.MustCompile(`(?i)Apple.*Fitness\+`)},
	{"Peloton", "Fitness & Bien-être", regexp.MustCompile(`(?i)Peloton`)},
	{"Headspace", "Fitness & Bien-être", regexp.MustCompile(`(?i)Headspace`)},
	{"Calm", "Fitness & Bien-être", regexp.MustCompile(`(?i)Calm`)},
	{"Nike Training Club", "Fitness & Bien-être", regexp.MustCompile(`(?i)Nike.*Training.*Club`)},
	{"LesMills+", "Fitness & Bien-être", regexp.MustCompile(`(?i)LesMills\+`)},
	{"WW (Weight Watchers)", "Fitness & Bien-être", regexp.MustCompile(`(?i)WW.*\(Weight.*Watchers\)`)},
	{"8fit", "Fitness & Bien-être", regexp.MustCompile(`(?i)8fit`)},
	{"Sweat", "Fitness & Bien-être", regexp.MustCompile(`(?i)Sweat`)},
	{"Alo Moves", "Fitness & Bien-être", regexp.MustCompile(`(?i)Alo.*Moves`)},
	{"Glo Yoga", "Fitness & Bien-être", regexp.MustCompile(`(?i)Glo.*Yoga`)},
	{"Nike Run Club Premium", "Fitness & Bien-être", regexp.MustCompile(`(?i)Nike.*Run.*Club.*Premium`)},

	// Éducation
	{"Udemy", "Éducation", regexp.MustCompile(`(?i)Udemy`)},
	{"Coursera", "Éducation", regexp.MustCompile(`(?i)Coursera`)},
	{"MasterClass", "Éducation", regexp.MustCompile(`(?i)MasterClass`)},
	{"Skillshare", "Éducation", regexp.MustCompile(`(?i)Skillshare`)},
	{"LinkedIn Learning", "Éducation", regexp.MustCompile(`(?i)LinkedIn.*Learning`)},
	{"Duolingo Plus", "Éducation", regexp.MustCompile(`(?i)Duolingo.*Plus`)},
	{"Brilliant", "Éducation", regexp.MustCompile(`(?i)Brilliant`)},
	{"Khan Academy", "Éducation", regexp.MustCompile(`(?i)Khan.*Academy`)},
	{"Codecademy", "Éducation", regexp.MustCompile(`(?i)Codecademy`)},
	{"Rosetta Stone", "Éducation", regexp.MustCompile(`(?i)Rosetta.*Stone`)},
	{"DataCamp", "Éducation", regexp.MustCompile(`(?i)DataCamp`)},
	{"edX Premium", "Éducation", regexp.MustCompile(`(?i)edX.*Premium`)},
	{"Memrise Pro", "Éducation", regexp.MustCompile(`(?i)Memrise.*Pro`)},
	{"SoloLearn Pro", "Éducation", regexp.MustCompile(`(?i)SoloLearn.*Pro`)},
	{"Pluralsight", "Éducation", regexp.MustCompile(`(?i)Pluralsight`)},

	// Cybersécurité
	{"NordVPN", "CyberSécurité", regexp.MustCompile(`(?i)NordVPN`)},
	{"ExpressVPN", "CyberSécurité", regexp.MustCompile(`(?i)ExpressVPN`)},
	{"Surfshark", "CyberSécurité", regexp.MustCompile(`(?i)Surfshark`)},
	{"ProtonVPN", "CyberSécurité", regexp.MustCompile(`(?i)ProtonVPN`)},
	{"CyberGhost", "CyberSécurité", regexp.MustCompile(`(?i)CyberGhost`)},
	{"1Password", "CyberSécurité", regexp.MustCompile(`(?i)1Password`)},
	{"LastPass", "CyberSécurité", regexp.MustCompile(`(?i)LastPass`)},
	{"Dashlane", "CyberSécurité", regexp.MustCompile(`(?i)Dashlane`)},
	{"ProtonMail", "CyberSécurité", regexp.MustCompile(`(?i)ProtonMail`)},
	{"Malwarebytes Premium", "CyberSécurité", regexp.MustCompile(`(?i)Malwarebytes.*Premium`)},
	{"TunnelBear VPN", "CyberSécurité", regexp.MustCompile(`(?i)TunnelBear.*VPN`)},
	{"Proton Pass", "CyberSécurité", regexp.MustCompile(`(?i)Proton.*Pass`)},
	{"Bitwarden Premium", "CyberSécurité", regexp.MustCompile(`(?i)Bitwarden.*Premium`)},

	// E-commerce & Livraisons
	{"Amazon Prime", "E-commerce & Livraisons", regexp.MustCompile(`(?i)Amazon.*Prime`)},
	{"Walmart+", "E-commerce & Livraisons", regexp.MustCompile(`(?i)Walmart\+`)},
	{"Shopify", "E-commerce & Livraisons", regexp.MustCompile(`(?i)Shopify`)},
	{"Deliveroo Plus", "E-commerce & Livraisons", regexp.MustCompile(`(?i)Deliveroo.*Plus`)},
	{"Uber One", "E-commerce & Livraisons", regexp.MustCompile(`(?i)Uber.*One`)},

	// Mode et Beauté
	{"Sephora Flash", "Mode & Beauté", regexp.MustCompile(`(?i)Sephora.*Flash`)},
	{"Zalando Plus", "Mode & Beauté", regexp.MustCompile(`(?i)Zalando.*Plus`)},
	{"H&M Membership", "Mode & Beauté", regexp.MustCompile(`(?i)H&M.*Membership`)},
	{"ASOS Premier", "Mode & Beauté", regexp.MustCompile(`(?i)ASOS.*Premier`)},
	{"Nike Membership", "Mode & Beauté", regexp.MustCompile(`(?i)Nike.*Membership`)},
	{"Beauty Pie", "Mode & Beauté", regexp.MustCompile(`(?i)Beauty.*Pie`)},
}
