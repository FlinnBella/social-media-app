package internal

const (
	corporate   = "corporate"
	upbeat      = "upbeat"
	realtor     = "realtor"
	traditional = "traditional"
)

// MusicThemeMap maps theme names to arrays of music filenames
var MusicThemeMap = map[string][]string{
	corporate: {
		"Aurora%20on%20the%20Boulevard%20-%20National%20Sweetheart.mp3",
		"Champion%20-%20Telecasted.mp3",
		"Crystaline%20-%20Quincas%20Moreira.mp3",
		"Final%20Soliloquy%20-%20Asher%20Fulero.mp3",
		"Hopeful%20-%20Nat%20Keefe.mp3",
		"Hopeful%20Freedom%20-%20Asher%20Fulero.mp3",
		"Name%20The%20Time%20And%20Place%20-%20Telecasted.mp3",
		"Organic%20Guitar%20House%20-%20Dyalla.mp3",
		"Phantom%20-%20Density%20%26%20Time.mp3",
		"Touch%20-%20Anno%20Domini%20Beats.mp3",
		"Traversing%20-%20Godmode.mp3",
	},
	upbeat: {
		"Baby%20Animals%20Playing%20-%20Joel%20Cummins.mp3",
		"Banjo%20Doops%20-%20Joel%20Cummins.mp3",
		"Buckle%20Up%20-%20Jeremy%20Korpas.mp3",
		"Cafecito%20por%20la%20Manana%20-%20Cumbia%20Deli.mp3",
		"Jetski%20-%20Telecasted.mp3",
		"Like%20It%20Loud%20-%20Dyalla.mp3",
		"Oh%20Please%20-%20Telecasted.mp3",
		"Seagull%20-%20Telecasted.mp3",
		"Sly%20Sky%20-%20Telecasted.mp3",
		"Twin%20Engines%20-%20Jeremy%20Korpas.mp3",
	},
	realtor: {
		"Heartbeat%20Of%20The%20Wind%20-%20Asher%20Fulero.mp3",
		"Hopeful%20-%20Nat%20Keefe.mp3",
		"Hopeful%20Freedom%20-%20Asher%20Fulero.mp3",
		"No.2%20Remembering%20Her%20-%20Esther%20Abrami.mp3",
		"Organic%20Guitar%20House%20-%20Dyalla.mp3",
		"Phantom%20-%20Density%20%26%20Time.mp3",
		"Touch%20-%20Anno%20Domini%20Beats.mp3",
		"Traversing%20-%20Godmode.mp3",
	},
	traditional: {
		"Curse%20of%20the%20Witches%20-%20Jimena%20Contreras.mp3",
		"Delayed%20Baggage%20-%20Ryan%20Stasik.mp3",
		"Honey%2C%20I%20Dismembered%20The%20Kids%20-%20Ezra%20Lipp.mp3",
		"Hopeless%20-%20Jimena%20Contreras.mp3",
		"Night%20Hunt%20-%20Jimena%20Contreras.mp3",
		"On%20The%20Hunt%20-%20Andrew%20Langdon.mp3",
		"Restless%20Heart%20-%20Jimena%20Contreras.mp3",
		"Sinister%20-%20Anno%20Domini%20Beats.mp3",
	},
}
