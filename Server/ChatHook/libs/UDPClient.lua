---@diagnostic disable: undefined-global

-- Made by Neverless @ BeamMP. Issues? Feel free to ask.
--[[
	Windows/Linux must unfortunately be separated atm.

	io.popen takes exceptionally long to execute on Linux for some reason.
	So on Windows the data is written to stdin, while on Linux its given as an argument.

	Use b64 encoding.
]]

local Log = require("libs/Log")
local RSocket

local SCRIPT_REF = "libUDPClient"

---@class ChatHookSocket
---@field _os string
---@field _bin string
---@field _ip string
---@field _port integer
---@field _client any|nil
---@field send fun(self: ChatHookSocket, data: string)
---@field close fun(self: ChatHookSocket)

---@param socket ChatHookSocket
---@return file*|nil
local function getUdpHandle(socket)
	return io.popen(string.format(
		"%s %s %s",
		socket._bin,
		socket._ip,
		socket._port
	), "w")
end

---@param socket ChatHookSocket
---@param data string
local function sendWindows(socket, data)
	local handle = getUdpHandle(socket)
	if not handle then
		return
	end

	handle:write(data)
	handle:close()
end

---@param socket ChatHookSocket
---@param data string
local function sendLinux(socket, data)
	if socket._client then
		socket._client:send(data)
		return
	end

	os.execute(string.format("%s %s %s %s", socket._bin, socket._ip, socket._port, data))
end

---@param osName string
---@param binPath string
---@return string|nil
local function normalizeBinPath(osName, binPath)
	if osName == "Linux" then
		return binPath
	end
	if osName == "Windows" then
		return binPath:gsub("%/", "\\")
	end
	return nil
end

---@param binPath string
---@param ip string
---@param port integer
---@return ChatHookSocket|nil
return function(binPath, ip, port)
	local osName = MP.GetOSName()
	binPath = normalizeBinPath(osName, binPath)
	if not binPath or not FS.Exists(binPath) then
		return nil
	end

	local client
	if osName == "Linux" then
		local useLib, lib = pcall(require, "rsocket")
		if not useLib then
			Log.info("Optional rsocket module is unavailable. Using helper binary fallback.", SCRIPT_REF)
		else
			Log.load("Successfully loaded Go RSocket module", SCRIPT_REF)
			RSocket = lib
			local isOk, socket = pcall(RSocket.udpClient, ip, port)
			if not isOk then
				Log.error("Cannot create RSocket UDP client. Using helper binary fallback.", SCRIPT_REF)
			else
				Log.load("Using RSocket UDP client.", SCRIPT_REF)
				client = socket
			end
		end
	end

	---@type ChatHookSocket
	local udp = {
		_os = osName,
		_bin = binPath,
		_ip = ip,
		_port = port,
		_client = client,
	}

	function udp:send(data)
		if self._os == "Windows" then
			return sendWindows(self, data)
		end
		if self._os == "Linux" then
			return sendLinux(self, data)
		end
	end

	function udp:close()
		if self._client then
			self._client:close()
		end
	end

	return udp
end
