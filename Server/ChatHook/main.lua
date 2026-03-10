-- Made by Neverless @ BeamMP. Issues? Feel free to ask.
---@diagnostic disable: lowercase-global

local VERSION = "0.11" -- 11.01.2026 (DD.MM.YYYY)

local DEFAULT_SCRIPT_REF = "ChatHook"
local DEBUG_SCRIPT_REF = "DebugHook"
local EVENT_TIMER_NAME = "chathook_bufprint"

package.loaded["libs/Build"] = nil
package.loaded["libs/Config"] = nil
package.loaded["libs/UDPClient"] = nil
package.loaded["libs/PlayerCount"] = nil
package.loaded["libs/colors"] = nil
package.loaded["libs/Log"] = nil

local Log = require("libs/Log").setCollectMode(true)
local Build = require("libs/Build")
local Config = require("libs/Config")
local UDPClient = require("libs/UDPClient")
local PlayerCount = require("libs/PlayerCount")
local Color = require("libs/colors")

local isDebugHook = false
local scriptRef = DEFAULT_SCRIPT_REF

local isStart = _G.IS_START == nil
_G.IS_START = true

---@class UdpSocket
---@field send fun(self: UdpSocket, data: string)
---@field close fun(self: UdpSocket)

---@type UdpSocket|nil
local socketClient = _G.CHATHOOK_SOCKET
if not isStart and socketClient then
	socketClient:close()
end
socketClient = nil
_G.CHATHOOK_SOCKET = nil

---@param fullPath string
---@return string|nil
local function filePath(fullPath)
	local _, pos = fullPath:find(".*/")
	if pos == nil then
		return nil
	end

	return fullPath:sub(1, pos)
end

---@return string
local function scriptPath()
	local sourcePath = debug.getinfo(2).source:gsub("\\", "/")
	if sourcePath:sub(1, 1) == "@" then
		return filePath(sourcePath:sub(2)) or ""
	end

	return filePath(sourcePath) or ""
end

---@class MessageBuffer
---@field _buf table[]
local Buf = { _buf = {} }

---@param build table
function Buf:add(build)
	table.insert(self._buf, build)
end

---@return integer
function Buf:size()
	return #self._buf
end

---@return table[]
function Buf:take()
	local ref = self._buf
	self._buf = {}
	return ref
end

function bufPrint()
	Log.printCollect()
	if socketClient and Buf:size() > 0 then
		socketClient:send(Build.wrap(Buf:take()))
	end
end

---@param playerId integer
---@param playerName string
---@param message string
function onChatMessage(playerId, playerName, message)
	local _ = playerName -- callback signature is fixed by BeamMP
	if message:len() == 0 or message:sub(1, 1) == "/" then
		return
	end

	Buf:add(Build.playerMessage(playerId, message))
end

---@param message string|nil
---@param callbackScriptRef string|nil
function onScriptMessage(message, callbackScriptRef)
	if message == nil or message:len() == 0 then
		return
	end

	if not isDebugHook then
		Buf:add(Build.scriptMessage(callbackScriptRef, message))
	else
		Buf:add(Build.scriptMessageNoBuf(callbackScriptRef, message))
	end
end

---@param playerId integer
function onPlayerConnecting(playerId)
	Buf:add(Build.playerJoining(playerId))
end

---@param playerId integer
function onPlayerJoin(playerId)
	PlayerCount.add(playerId)
	Buf:add(Build.playerJoin(playerId))
end

---@param playerId integer
function onPlayerDisconnect(playerId)
	local wasSynced = PlayerCount.remove(playerId)
	Buf:add(Build.playerLeft(playerId, not wasSynced))
end

function onInit()
	local configPath = scriptPath() .. "config.json"
	local config = Config.load(configPath, Log, scriptRef)
	if not config then
		return
	end

	isDebugHook = config.debugHook
	scriptRef = isDebugHook and DEBUG_SCRIPT_REF or DEFAULT_SCRIPT_REF

	MP.CancelEventTimer(EVENT_TIMER_NAME)
	MP.RegisterEvent(EVENT_TIMER_NAME, "bufPrint")
	MP.CreateEventTimer(EVENT_TIMER_NAME, config.flushIntervalMs)

	Log.load("====. Loading " .. scriptRef .. " .====", scriptRef)
	Log.load("> Version: " .. VERSION, scriptRef)
	Log.load("> Build Packet Version: " .. Build.VERSION, scriptRef)
	Log.info("^ ^n(This must match with the ChatHook daemon version)^r", scriptRef)

	Log.ok("> Server Name: " .. Color.convertToConsole(config.serverName), scriptRef)
	Log.ok("> Max Players: " .. config.maxPlayers, scriptRef)
	Build.setServerName(config.serverName)
	Build.setMaxPlayers(config.maxPlayers)

	local binPath = scriptPath() .. "bin/udp"
	local osName = MP.GetOSName()
	Log.info("> Operating System: " .. osName, scriptRef)

	if osName == "Windows" then
		Log.load("> Applying Windows patch", scriptRef)
		binPath = binPath .. ".exe"
	elseif osName == "Linux" then
		Log.load("> Applying Linux patch", scriptRef)
		os.execute("chmod +x \"" .. binPath .. "\"")
	else
		Log.fatal("Unsupported Operating System", scriptRef)
		return
	end

	if not FS.Exists(binPath) then
		Log.fatal("Cannot find udp binary in \"" .. binPath .. "\"", scriptRef)
		return
	end

	Log.load("> Building UDPSocket for " .. config.chatHookIp .. ":" .. config.udpPort, scriptRef)
	socketClient = UDPClient(binPath, config.chatHookIp, config.udpPort)
	if socketClient == nil then
		Log.fatal("Cannot create UDPSocket", scriptRef)
		return
	end

	_G.CHATHOOK_SOCKET = socketClient
	Log.ok("> Initialized UDPSocket", scriptRef)

	if not isDebugHook then
		MP.RegisterEvent("onChatMessage", "onChatMessage")
		MP.RegisterEvent("onPlayerConnecting", "onPlayerConnecting")
		MP.RegisterEvent("onPlayerJoin", "onPlayerJoin")
		MP.RegisterEvent("onPlayerDisconnect", "onPlayerDisconnect")
		MP.RegisterEvent("onScriptMessage", "onScriptMessage")
	else
		MP.RegisterEvent("onDebugMessage", "onScriptMessage")
	end

	if isStart then
		Buf:add(Build.serverOnline())
	else
		Buf:add(Build.serverReload())
	end

	for playerId, playerName in pairs(MP.GetPlayers() or {}) do
		PlayerCount.add(playerId)
		Log.ok("Hotreloaded \"" .. playerName .. "\"", scriptRef)
	end

	Log.load("=====. " .. scriptRef .. " Loaded .====", scriptRef)
	Log.printCollect()
end
