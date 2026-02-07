package format

import "fmt"

func Time(seconds float64) string {
	totalSeconds := int(seconds)
	minutes := totalSeconds / 60
	secs := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

func Model(model string, skin byte) string {
	switch model {
	case "maps/b_bh100.bsp":
		return "mega"
	case "progs/armor.mdl":
		switch skin {
		case 0:
			return "ga"
		case 1:
			return "ya"
		case 2:
			return "ra"
		default:
			return "unknown"
		}
	case "progs/g_shot.mdl":
		return "ssg"
	case "progs/g_nail.mdl":
		return "ng"
	case "progs/g_nail2.mdl":
		return "sng"
	case "progs/g_rock.mdl":
		return "gl"
	case "progs/g_rock2.mdl":
		return "rl"
	case "progs/g_light.mdl":
		return "lg"
	case "progs/quaddama.mdl":
		return "quad"
	case "progs/invulner.mdl":
		return "pent"
	case "progs/invisibl.mdl":
		return "ring"
	default:
		return model
	}
}
