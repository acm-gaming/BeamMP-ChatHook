//go:build lua_module

package main

/*
#cgo CFLAGS: -std=c11
#include <stdint.h>
#include <stdlib.h>

typedef struct lua_State lua_State;
typedef long long lua_Integer;
typedef int (*lua_CFunction)(lua_State *L);

typedef struct luaL_Reg {
	const char *name;
	lua_CFunction func;
} luaL_Reg;

enum {
	LUA_REGISTRYINDEX = -1001000,
};

extern int luaL_newmetatable(lua_State *L, const char *tname);
extern void luaL_setfuncs(lua_State *L, const luaL_Reg *l, int nup);
extern void *lua_newuserdata(lua_State *L, size_t sz);
extern void *lua_touserdata(lua_State *L, int idx);
extern void lua_pushvalue(lua_State *L, int idx);
extern void lua_setfield(lua_State *L, int idx, const char *k);
extern void lua_getfield(lua_State *L, int idx, const char *k);
extern void lua_setmetatable(lua_State *L, int objindex);
extern void lua_pushcclosure(lua_State *L, lua_CFunction fn, int n);
extern void lua_createtable(lua_State *L, int narr, int nrec);
extern void lua_settop(lua_State *L, int idx);
extern const char *luaL_checkstring(lua_State *L, int arg);
extern lua_Integer luaL_checkinteger(lua_State *L, int arg);
extern void lua_pushnil(lua_State *L);
extern void lua_pushlstring(lua_State *L, const char *s, size_t len);
extern void lua_pushstring(lua_State *L, const char *s);
extern int luaL_error(lua_State *L, const char *fmt, ...);

extern int go_lua_new_udp_client(lua_State *L);
extern int go_lua_new_udp_listener(lua_State *L);
extern int go_lua_client_send(lua_State *L);
extern int go_lua_client_recv_string(lua_State *L);
extern int go_lua_client_close(lua_State *L);
extern int go_lua_listener_send_to(lua_State *L);
extern int go_lua_listener_recv_string(lua_State *L);
extern int go_lua_listener_close(lua_State *L);

static int wrap_new_udp_client(lua_State *L) { return go_lua_new_udp_client(L); }
static int wrap_new_udp_listener(lua_State *L) { return go_lua_new_udp_listener(L); }
static int wrap_client_send(lua_State *L) { return go_lua_client_send(L); }
static int wrap_client_recv_string(lua_State *L) { return go_lua_client_recv_string(L); }
static int wrap_client_close(lua_State *L) { return go_lua_client_close(L); }
static int wrap_listener_send_to(lua_State *L) { return go_lua_listener_send_to(L); }
static int wrap_listener_recv_string(lua_State *L) { return go_lua_listener_recv_string(L); }
static int wrap_listener_close(lua_State *L) { return go_lua_listener_close(L); }

static const luaL_Reg client_methods[] = {
	{"send", wrap_client_send},
	{"recvString", wrap_client_recv_string},
	{"close", wrap_client_close},
	{"__gc", wrap_client_close},
	{NULL, NULL},
};

static const luaL_Reg listener_methods[] = {
	{"sendTo", wrap_listener_send_to},
	{"recvString", wrap_listener_recv_string},
	{"close", wrap_listener_close},
	{"__gc", wrap_listener_close},
	{NULL, NULL},
};

static void register_metatable(lua_State *L, const char *name, const luaL_Reg *methods) {
	luaL_newmetatable(L, name);
	lua_pushvalue(L, -1);
	lua_setfield(L, -2, "__index");
	luaL_setfuncs(L, methods, 0);
	lua_settop(L, -2);
}

static void set_named_metatable(lua_State *L, const char *name) {
	lua_getfield(L, LUA_REGISTRYINDEX, name);
	lua_setmetatable(L, -2);
}

static int push_error(lua_State *L, const char *msg) {
	return luaL_error(L, "%s", msg);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

type handleBox struct {
	handle uint64
}

var handleStore sync.Map

func storeHandle(value any) uint64 {
	handle := nextHandle.Add(1)
	handleStore.Store(handle, value)
	return handle
}

func loadHandle[T any](box *handleBox) (*T, bool) {
	if box == nil || box.handle == 0 {
		return nil, false
	}
	value, ok := handleStore.Load(box.handle)
	if !ok {
		return nil, false
	}
	typed, ok := value.(*T)
	return typed, ok
}

func deleteHandle(box *handleBox) {
	if box == nil || box.handle == 0 {
		return
	}
	handleStore.Delete(box.handle)
	box.handle = 0
}

func newHandleUserdata(L *C.lua_State, metatableName string, value any) *handleBox {
	box := (*handleBox)(C.lua_newuserdata(L, C.size_t(unsafe.Sizeof(handleBox{}))))
	box.handle = storeHandle(value)
	name := C.CString(metatableName)
	defer C.free(unsafe.Pointer(name))
	C.set_named_metatable(L, name)
	return box
}

func requireClient(L *C.lua_State) (*udpClient, *handleBox, bool) {
	box := (*handleBox)(C.lua_touserdata(L, 1))
	client, ok := loadHandle[udpClient](box)
	if !ok {
		return nil, nil, false
	}
	return client, box, true
}

func requireListener(L *C.lua_State) (*udpListener, *handleBox, bool) {
	box := (*handleBox)(C.lua_touserdata(L, 1))
	listener, ok := loadHandle[udpListener](box)
	if !ok {
		return nil, nil, false
	}
	return listener, box, true
}

func pushString(L *C.lua_State, value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.lua_pushlstring(L, cValue, C.size_t(len(value)))
}

func luaError(L *C.lua_State, err error) C.int {
	message := C.CString(err.Error())
	defer C.free(unsafe.Pointer(message))
	return C.push_error(L, message)
}

func registerModule(L *C.lua_State) {
	clientName := C.CString("rsocket.udpClient")
	listenerName := C.CString("rsocket.udpListener")
	defer C.free(unsafe.Pointer(clientName))
	defer C.free(unsafe.Pointer(listenerName))

	C.register_metatable(L, clientName, &C.client_methods[0])
	C.register_metatable(L, listenerName, &C.listener_methods[0])

	C.lua_createtable(L, 0, 2)
	udpClientKey := C.CString("udpClient")
	udpListenerKey := C.CString("udpListener")
	defer C.free(unsafe.Pointer(udpClientKey))
	defer C.free(unsafe.Pointer(udpListenerKey))

	C.lua_pushcclosure(L, (C.lua_CFunction)(C.wrap_new_udp_client), 0)
	C.lua_setfield(L, -2, udpClientKey)
	C.lua_pushcclosure(L, (C.lua_CFunction)(C.wrap_new_udp_listener), 0)
	C.lua_setfield(L, -2, udpListenerKey)
}

//export luaopen_rsocket
func luaopen_rsocket(L *C.lua_State) C.int {
	registerModule(L)
	return 1
}

//export go_lua_new_udp_client
func go_lua_new_udp_client(L *C.lua_State) C.int {
	ip := C.GoString(C.luaL_checkstring(L, 1))
	port := int(C.luaL_checkinteger(L, 2))

	client, err := newUDPClient(ip, port)
	if err != nil {
		return luaError(L, fmt.Errorf("cannot connect to given address %s:%d", ip, port))
	}

	newHandleUserdata(L, "rsocket.udpClient", client)
	return 1
}

//export go_lua_new_udp_listener
func go_lua_new_udp_listener(L *C.lua_State) C.int {
	ip := C.GoString(C.luaL_checkstring(L, 1))
	port := int(C.luaL_checkinteger(L, 2))

	listener, err := newUDPListener(ip, port)
	if err != nil {
		return luaError(L, fmt.Errorf("cannot bind to given address %s:%d", ip, port))
	}

	newHandleUserdata(L, "rsocket.udpListener", listener)
	return 1
}

//export go_lua_client_send
func go_lua_client_send(L *C.lua_State) C.int {
	client, _, ok := requireClient(L)
	if !ok {
		return luaError(L, errors.New("cannot send data"))
	}

	data := C.GoString(C.luaL_checkstring(L, 2))
	if err := client.Send(data); err != nil {
		return luaError(L, err)
	}

	return 0
}

//export go_lua_client_recv_string
func go_lua_client_recv_string(L *C.lua_State) C.int {
	client, _, ok := requireClient(L)
	if !ok {
		return luaError(L, errors.New("cannot receive data"))
	}

	value, okValue, err := client.RecvString()
	if err != nil {
		return luaError(L, err)
	}
	if !okValue {
		C.lua_pushnil(L)
		return 1
	}

	pushString(L, value)
	return 1
}

//export go_lua_client_close
func go_lua_client_close(L *C.lua_State) C.int {
	client, box, ok := requireClient(L)
	if ok {
		_ = client.Close()
		deleteHandle(box)
	}
	return 0
}

//export go_lua_listener_send_to
func go_lua_listener_send_to(L *C.lua_State) C.int {
	listener, _, ok := requireListener(L)
	if !ok {
		return luaError(L, errors.New("cannot send data"))
	}

	data := C.GoString(C.luaL_checkstring(L, 2))
	ip := C.GoString(C.luaL_checkstring(L, 3))
	port := int(C.luaL_checkinteger(L, 4))

	if err := listener.SendTo(data, ip, port); err != nil {
		return luaError(L, err)
	}

	return 0
}

//export go_lua_listener_recv_string
func go_lua_listener_recv_string(L *C.lua_State) C.int {
	listener, _, ok := requireListener(L)
	if !ok {
		return luaError(L, errors.New("cannot receive data"))
	}

	value, source, okValue, err := listener.RecvString()
	if err != nil {
		return luaError(L, err)
	}
	if !okValue {
		C.lua_pushnil(L)
		C.lua_pushnil(L)
		return 2
	}

	pushString(L, value)
	pushString(L, source)
	return 2
}

//export go_lua_listener_close
func go_lua_listener_close(L *C.lua_State) C.int {
	listener, box, ok := requireListener(L)
	if ok {
		_ = listener.Close()
		deleteHandle(box)
	}
	return 0
}
