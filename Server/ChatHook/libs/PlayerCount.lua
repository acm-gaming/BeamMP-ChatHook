---@diagnostic disable: undefined-global

local M = {}

---@type table<integer, boolean>
local players = {}

---@param playerId integer
function M.add(playerId)
	players[playerId] = true
end

---@param playerId integer
---@return boolean?
function M.remove(playerId)
	if not M.exists(playerId) then
		return
	end
	players[playerId] = nil
	return true
end

---@param playerId integer
---@return boolean
function M.exists(playerId)
	return players[playerId] ~= nil
end

---@return integer
function M.diff()
	local count = M.count()
	if count == 0 then
		return 0
	end
	local total = 0
	for _, _ in pairs(MP.GetPlayers() or {}) do
		total = total + 1
	end
	return total - count
end

---@return integer
function M.count()
	local total = 0
	for _, _ in pairs(players) do
		total = total + 1
	end
	return total
end

return M
