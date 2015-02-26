// liteapp.go
package main

/*
extern void cdrv_init(void *fn);
extern int cdrv_main(int argc,char** argv);
//extern void cdrv_cb(void *cb, void *id, void *reply, int size, int err, void* ctx);

static void cdrv_init_ex()
{
	extern int godrv_call(void* id,int id_size, void* args, int args_size, void* cb, void* ctx);
	cdrv_init(&godrv_call);
}

static void cdrv_cb(void *cb, void *id, int id_size, void *reply, int size, int err, void* ctx)
{
	typedef void (*DRV_CALLBACK)(void *id, int id_size, void *reply, int len, int err, void *ctx);
    ((DRV_CALLBACK)(cb))(id,id_size,reply,size,err,ctx);
}

#cgo windows LDFLAGS: -L../../liteide/bin -lliteapp
#cgo linux LDFLAGS: -L../../liteide/bin -lliteapp
#cgo openbsd LDFLAGS: -L../../liteide/bin -lliteapp
#cgo darwin LDFLAGS: -L../../liteide/bin/liteide.app/Contents/MacOS
*/
import "C"
import (
	"log"
	"strings"
	"unsafe"

	"github.com/visualfc/gotools/command"
)

func cdrv_main(args []string) int {
	argc := len(args)
	var cargs []*C.char
	for _, arg := range args {
		size := len(arg)
		data := make([]C.char, size+1)
		for i := 0; i < size; i++ {
			data[i] = (C.char)(arg[i])
		}
		data[size] = 0
		cargs = append(cargs, &data[0])
	}
	C.cdrv_init_ex()
	return int(C.cdrv_main(C.int(argc), &cargs[0]))
}

func cdrv_cb(cb unsafe.Pointer, id []byte, reply []byte, err int, ctx unsafe.Pointer) {
	C.cdrv_cb(cb, unsafe.Pointer(&id[0]), C.int(len(id)), unsafe.Pointer(&reply[0]), C.int(len(reply)), C.int(err), ctx)
}

//export godrv_call
func godrv_call(id unsafe.Pointer, id_size C.int, args unsafe.Pointer, size C.int, cb unsafe.Pointer, ctx unsafe.Pointer) C.int {
	return C.int(go_call(C.GoBytes(id, id_size), C.GoBytes(args, size), cb, ctx))
}

type writer_output struct {
	id  []byte
	cb  unsafe.Pointer
	ctx unsafe.Pointer
	err int
}

func (w *writer_output) Write(p []byte) (n int, err error) {
	n = len(p)
	log.Println(string(w.id), string(p))
	cdrv_cb(w.cb, w.id, p, w.err, w.ctx)
	return
}

func go_call(id []byte, args []byte, cb unsafe.Pointer, ctx unsafe.Pointer) int {
	for _, cmd := range command.CommandList() {
		if cmd == string(id) {
			go func() {
				command.Stdout = &writer_output{id, cb, ctx, 0}
				command.Stderr = &writer_output{id, cb, ctx, -1}
				var arguments []string
				arguments = append(arguments, string(id))
				if len(args) > 0 {
					for _, opt := range strings.Split(string(args), ";;") {
						if len(opt) > 0 {
							arguments = append(arguments, opt)
						}
					}
				}
				command.ParseArgs(arguments)
			}()
			return 0
		}
	}
	//	return 1
	//	if fn, ok := cmdFuncMap[string(id)]; ok {
	//		go func(id, args []byte, cb, ctx unsafe.Pointer) {
	//			rep, err := fn(args)
	//			if err != nil {
	//				cdrv_cb(cb, id, []byte{0}, -1, ctx)
	//			}
	//			cdrv_cb(cb, id, append(rep, 0), 0, ctx)
	//		}(id, args, cb, ctx)
	//		return 0
	//	}
	return -1
}
