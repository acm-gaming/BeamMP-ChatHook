---@diagnostic disable: undefined-global

-- Made by Neverless @ BeamMP. Issues? Feel free to ask.
local Base64 = require("libs/base64")
local PlayerCount = require("libs/PlayerCount")

local M = {
	SERVER_NAME = "",
	MAX_PLAYERS = 0,
	VERSION = 4,
}

---@param payload table
---@return string
local function wrap(payload)
	return Base64.encode(Util.JsonEncode(payload))
end

---@return table
local function baseBuild()
	return {
		server_name = M.SERVER_NAME,
		player_count = PlayerCount.count(),
		player_dif = PlayerCount.diff(),
		player_max = M.MAX_PLAYERS,
		version = M.VERSION,
	}
end

---@param payload table
---@return table
local function withBase(payload)
	for key, value in pairs(baseBuild()) do
		payload[key] = value
	end

	return payload
end

---@param serverName string
function M.setServerName(serverName)
	M.SERVER_NAME = serverName
end

---@param maxPlayers integer
function M.setMaxPlayers(maxPlayers)
	M.MAX_PLAYERS = maxPlayers
end

---@param contents table[]
---@return string
function M.wrap(contents)
	return wrap(withBase({
		contents = contents,
	}))
end

---@param playerId integer
---@param message string
---@return table
function M.playerMessage(playerId, message)
	return {
		type = 1,
		content = {
			player_name = MP.GetPlayerName(playerId),
			chat_message = message,
		},
	}
end

---@param messageScriptRef string|nil
---@param message string
---@return table
function M.scriptMessage(messageScriptRef, message)
	return {
		type = 6,
		content = {
			script_ref = messageScriptRef or "",
			chat_message = message,
		},
	}
end

---@return table
function M.serverOnline()
	return {
		type = 2,
	}
end

---@return table
function M.serverReload()
	return {
		type = 5,
	}
end

---@param playerId integer
---@return table
function M.playerJoining(playerId)
	return {
		type = 7,
		content = {
			player_name = MP.GetPlayerName(playerId),
		},
	}
end

---@param playerId integer
---@return table
function M.playerJoin(playerId)
	return {
		type = 3,
		content = {
			player_name = MP.GetPlayerName(playerId),
			ip = MP.GetPlayerIdentifiers(playerId).ip,
		},
	}
end

---@param playerId integer
---@param early boolean
---@return table
function M.playerLeft(playerId, early)
	return {
		type = 4,
		content = {
			player_name = MP.GetPlayerName(playerId),
			early = early,
		},
	}
end

return M
