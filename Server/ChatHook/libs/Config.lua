---@diagnostic disable: undefined-global

---@class ChatHookConfig
---@field serverName string
---@field maxPlayers integer
---@field chatHookIp string
---@field udpPort integer
---@field debugHook boolean
---@field flushIntervalMs integer

local Config = {}

---@param value string
---@return string
local function trim(value)
	return value:match("^%s*(.-)%s*$")
end

---@param path string
---@return string|nil, string?
local function readFile(path)
	local handle, openError = io.open(path, "r")
	if handle == nil then
		return nil, string.format('Cannot open "%s": %s', path, openError or "unknown error")
	end

	local contents = handle:read("*a")
	handle:close()
	if contents == nil or contents == "" then
		return nil, string.format('Config file "%s" is empty', path)
	end

	return contents
end

---@param rawJson string
---@return table|nil, string?
local function decodeJson(rawJson)
	local isOk, decoded = pcall(Util.JsonDecode, rawJson)
	if not isOk then
		return nil, "config.json contains invalid JSON"
	end
	if type(decoded) ~= "table" then
		return nil, "config.json must decode to a JSON object"
	end

	return decoded
end

---@param config table
---@param key string
---@return string|nil, string?
local function readString(config, key)
	local value = config[key]
	if type(value) ~= "string" then
		return nil, string.format('config.%s must be a string', key)
	end

	value = trim(value)
	if value == "" then
		return nil, string.format('config.%s cannot be empty', key)
	end

	return value
end

---@param config table
---@param key string
---@param minimum integer
---@param maximum integer
---@return integer|nil, string?
local function readInteger(config, key, minimum, maximum)
	local value = config[key]
	if type(value) ~= "number" or value % 1 ~= 0 then
		return nil, string.format('config.%s must be an integer', key)
	end
	if value < minimum or value > maximum then
		return nil, string.format('config.%s must be between %d and %d', key, minimum, maximum)
	end

	return value
end

---@param config table
---@param key string
---@param default boolean
---@return boolean|nil, string?
local function readBoolean(config, key, default)
	local value = config[key]
	if value == nil then
		return default
	end
	if type(value) ~= "boolean" then
		return nil, string.format('config.%s must be a boolean', key)
	end

	return value
end

---@param config table
---@param key string
---@param default integer
---@param minimum integer
---@param maximum integer
---@return integer|nil, string?
local function readOptionalInteger(config, key, default, minimum, maximum)
	local value = config[key]
	if value == nil then
		return default
	end
	if type(value) ~= "number" or value % 1 ~= 0 then
		return nil, string.format('config.%s must be an integer', key)
	end
	if value < minimum or value > maximum then
		return nil, string.format('config.%s must be between %d and %d', key, minimum, maximum)
	end

	return value
end

---@param rawConfig table
---@return ChatHookConfig|nil, string?
local function normalizeConfig(rawConfig)
	local serverName, serverNameError = readString(rawConfig, "serverName")
	if serverName == nil then
		return nil, serverNameError
	end

	local maxPlayers, maxPlayersError = readInteger(rawConfig, "maxPlayers", 1, 1024)
	if maxPlayers == nil then
		return nil, maxPlayersError
	end

	local chatHookIp, chatHookIpError = readString(rawConfig, "chatHookIp")
	if chatHookIp == nil then
		return nil, chatHookIpError
	end

	local udpPort, udpPortError = readInteger(rawConfig, "udpPort", 1, 65535)
	if udpPort == nil then
		return nil, udpPortError
	end

	local debugHook, debugHookError = readBoolean(rawConfig, "debugHook", false)
	if debugHook == nil then
		return nil, debugHookError
	end

	local flushIntervalMs, flushIntervalError = readOptionalInteger(rawConfig, "flushIntervalMs", 1000, 250, 10000)
	if flushIntervalMs == nil then
		return nil, flushIntervalError
	end

	return {
		serverName = serverName,
		maxPlayers = maxPlayers,
		chatHookIp = chatHookIp,
		udpPort = udpPort,
		debugHook = debugHook,
		flushIntervalMs = flushIntervalMs,
	}
end

---@param path string
---@param log table
---@param scriptRef string
---@return ChatHookConfig|nil
function Config.load(path, log, scriptRef)
	local raw, readError = readFile(path)
	if raw == nil then
		log.fatal(readError or ("Cannot read config file at " .. path), scriptRef)
		return nil
	end

	local decoded, decodeError = decodeJson(raw)
	if decoded == nil then
		log.fatal(decodeError or "Cannot parse config.json", scriptRef)
		return nil
	end

	local normalized, validationError = normalizeConfig(decoded)
	if normalized == nil then
		log.fatal(validationError or "Invalid config.json", scriptRef)
		return nil
	end

	return normalized
end

return Config
