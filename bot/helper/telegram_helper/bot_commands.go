package telegram_helper

type BotCommandsType struct {
	StartCommand       string
	MirrorCommand      []string
	ZipMirrorCommand   []string
	UnzipMirrorCommand []string
	LeechCommand       []string
	ZipLeechCommand    []string
	UnzipLeechCommand  []string
	YtdlCommand        []string
	YtdlZipCommand     []string
	YtdlLeechCommand   []string
	YtdlZipLeech       []string
	CloneCommand       []string
	CountCommand       string
	CancelCommand      []string
	CancelAllCommand   string
	StatusCommand      []string
	StatsCommand       []string
	PingCommand        []string
	HelpCommand        []string
	AuthorizeCommand   []string
	UnAuthorizeCommand []string
	AuthListCommand    string
	ShellCommand       string
	SpeedCommand       []string
	MediaInfoCommand   []string
	UserSetCommand     []string
	BroadcastCommand   []string
	RestartCommand     []string
	LogCommand         string
}

var BotCommands = BotCommandsType{
	StartCommand:       "/start",
	MirrorCommand:      []string{"/mirror", "/m"},
	ZipMirrorCommand:   []string{"/zipmirror", "/zm"},
	UnzipMirrorCommand: []string{"/unzipmirror", "/uzm"},
	LeechCommand:       []string{"/leech", "/l"},
	ZipLeechCommand:    []string{"/zipleech", "/zl"},
	UnzipLeechCommand:  []string{"/unzipleech", "/uzl"},
	YtdlCommand:        []string{"/ytdl", "/y"},
	YtdlZipCommand:     []string{"/ytdlzip", "/yz"},
	YtdlLeechCommand:   []string{"/ytdlleech", "/yl"},
	YtdlZipLeech:       []string{"/ytdlzipleech", "/yzl"},
	CloneCommand:       []string{"/clone", "/c"},
	CountCommand:       "/count",
	CancelCommand:      []string{"/cancel", "/stop"},
	CancelAllCommand:   "/cancelall",
	StatusCommand:      []string{"/status", "/s", "/statusall"},
	StatsCommand:       []string{"/stats", "/st"},
	PingCommand:        []string{"/ping", "/p"},
	HelpCommand:        []string{"/help", "/h"},
	AuthorizeCommand:   []string{"/authorize", "/a", "/auth"},
	UnAuthorizeCommand: []string{"/unauthorize", "/ua", "/unauth"},
	AuthListCommand:    "/authlist",
	ShellCommand:       "/shell",
	SpeedCommand:       []string{"/speedtest", "/sp"},
	MediaInfoCommand:   []string{"/mediainfo", "/mi"},
	UserSetCommand:     []string{"/usersettings", "/usetting", "/us"},
	BroadcastCommand:   []string{"/broadcast", "/bc"},
	RestartCommand:     []string{"/restart", "/r", "/restartall"},
	LogCommand:         "/log",
}
