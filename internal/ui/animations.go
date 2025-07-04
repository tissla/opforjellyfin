package ui

//emojistore:
// 📁 🎞️ 🌐 🗃️ 🕵️ 🗄️ ➝ ⏳

// animation map
// Spinners
var Animations = map[string][]string{
	"MoviePlacement": MoviePlacement,
	"MetaFetcher":    MetaFetcher,
	"Searcher":       Searcher,
	"DownloadPrep":   DownloadPrep,
}

// cool animations
var MoviePlacement = []string{
	"📁🎞️ ➝          📂",
	"📁 🎞️ ➝         📂",
	"📁  🎞️ ➝        📂",
	"📁   🎞️ ➝       📂",
	"📁    🎞️ ➝      📂",
	"📁     🎞️ ➝     📂",
	"📁      🎞️ ➝    📂",
	"📁       🎞️ ➝   📂",
	"📁        🎞️ ➝  📂",
	"📁         🎞️ ➝ 📂",
	"📁          🎞️ ➝📂",
}

var MetaFetcher = []string{
	"🌐📁➝           🗃️",
	"🌐 📁➝          🗃️",
	"🌐  📁➝         🗃️",
	"🌐   📁➝        🗃️",
	"🌐    📁➝       🗃️",
	"🌐     📁➝      🗃️",
	"🌐      📁➝     🗃️",
	"🌐       📁➝    🗃️",
	"🌐        📁➝   🗃️",
	"🌐         📁➝  🗃️",
	"🌐          📁➝ 🗃️",
	"🌐           📁➝🗃️",
}

// extravaganza
var Searcher = []string{
	"     🕵️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🕵️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🕵️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🕵️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗃️🕵️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗃️🗃️🕵️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗃️🗃️🗃️🕵️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🕵️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🗄️🕵️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🗄️🕵️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗄️🕵️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🕵️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🕵️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🕵️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗄️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🕵️🗄️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗃️🕵️🗄️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗃️🗃️🕵️🗄️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗄️🗄️\n     🗄️🗄️🗄️🗃️🗃️🗃️🕵️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🕵️🗄️\n     🗄️🗄️🗄️🗃️🗃️🗃️🗃️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗃️🕵️\n     🗄️🗄️🗄️🗃️🗃️🗃️🗃️🗄️",
	"     🗃️🗃️🗃️🗃️🗃️🗃️🗃️🗄️\n     🗄️🗄️🗃️🗃️🗃️🗃️🗃️🕵️\n     🗄️🗄️🗄️🗃️🗄️🗄️🗃️🗃️\n     🗄️🗄️🗄️🗃️🗃️🗃️🗃️🗄️",
}

// unused i have no fantasy
var DownloadPrep = []string{
	"   🧍         \n              \n   🚀         ",
	"   🧍         \n    🧍        \n   🚀         ",
	"   🧍         \n    🧍🔧      \n   🚀         ",
	"   🧍         \n  🧍 🔧       \n   🚀         ",
	"   🧍         \n  🧍🔧        \n   🚀         ",
	"   🧍💬       \n  🧍✅        \n   🚀⚙️       ",
	"   👋         \n  🧍 ✅       \n   🚀🔥       ",
	"              \n  👋          \n   🚀🔥💨     ",
	"              \n              \n   🚀🔥🔥💨   ",
	"              \n              \n   🚀💨💨💨💨 ",
}
