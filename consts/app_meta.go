package consts

var (
	ServiceName  = "unnamed-service"
	ModuleName   = "unnamed-module"
	Version      = "0.0.0"
	VcsRevision  = "dirty"
	LibGoVersion = "0.0.0"
)

//func init() {
//	if name, ok := os.LookupEnv("SERVICE_NAME"); ok {
//		ServiceName = name
//	}
//	if name, ok := os.LookupEnv("MODULE_NAME"); ok {
//		ModuleName = name
//	}
//}
//
//func init() {
//	const libgoPath = "github.com/ca-media-nantes/libgo"
//	info, ok := debug.ReadBuildInfo()
//	if !ok {
//		return
//	}
//	for _, m := range info.Deps {
//		if strings.Contains(m.Path, libgoPath) {
//			LibGoVersion = m.Version
//			return
//		}
//	}
//}
