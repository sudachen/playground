package log

import(
	"github.com/ethereum/go-ethereum/log"
	"github.com/sudachen/misc/out"
	"fmt"
	"strings"
)

type lgx struct {}

func init() {
	log.Root().SetHandler(&lgx{})
}

func level(lvl log.Lvl) out.Level {
	switch lvl {
	case log.LvlCrit:  return out.Crit
	case log.LvlError: return out.Error
	case log.LvlWarn:  return out.Warn
	case log.LvlInfo:  return out.Info
	case log.LvlDebug: return out.Debug
	case log.LvlTrace: return out.Trace
	}
	return out.Trace
}

func (l *lgx) Log(r *log.Record) error {
	a := make([]string,len(r.Ctx)+1)
	a[0] = r.Msg
	for i, x := range r.Ctx {
		a[i+1] = fmt.Sprint(x)
	}
	level(r.Lvl).Print(strings.Join(a," "))
	return nil
}
