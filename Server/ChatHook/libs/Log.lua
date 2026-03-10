---@diagnostic disable: undefined-global

local Col = require("libs/colors")

local M = {}

local collectMode = false
local collectBuffer = ""

---@param value string
---@return string
local function fileName(value)
	local cleaned = value:sub(1):gsub("\\", "/")
	local _, pos = cleaned:find(".*/")
	if pos == nil then
		return value
	end
	return cleaned:sub(pos + 1, -1)
end

---@param file string
---@return string
local function cleanseName(file)
	local name = fileName(file)
	local final = name:find("%.")
	if final then
		final = final - 1
	end

	return name:sub(1, final)
end

---@param display string|nil
---@return string
local function stackTrace(display)
	local stackTraceOutput = "\n"
	local index = 3
	while debug.getinfo(index) do
		local info = debug.getinfo(index)
		local source = info.source or ""
		if source == "=[C]" then
			source = "builtin"
		end

		local name = info.name
		if name == nil then
			name = "-"
		else
			name = Col.ifServer(name .. "()", Col.orange)
		end

		local lineDefined = info.linedefined
		if lineDefined < 1 then
			lineDefined = "-"
		end

		local spacer = ""
		if index > 3 then
			spacer = " ^ "
		end

		stackTraceOutput = stackTraceOutput
			.. spacer
			.. Col.ifServer(fileName(source), Col.bold)
			.. "@"
			.. name
			.. ":"
			.. lineDefined
			.. "\n"

		index = index + 1
	end

	if display then
		if log then
			log(display, "== STACKTRACE ==", stackTraceOutput)
		else
			Col.print(stackTraceOutput, Col.bold("== STACKTRACE =="))
		end
	end

	return stackTraceOutput
end

---@param state boolean
---@return table
function M.setCollectMode(state)
	collectMode = state
	return M
end

function M.printCollect()
	local length = collectBuffer:len()
	if not collectMode or length == 0 then
		return
	end
	printRaw(collectBuffer:sub(1, length - 1))
	collectBuffer = ""
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.fatal(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("E")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("E", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.lightRed("FATAL")) .. "\n"
		else
			Col.print(display .. reason, Col.lightRed("FATAL"))
		end
	end
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.error(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("E")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("E", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.lightRed("ERROR")) .. "\n"
		else
			Col.print(display .. reason, Col.lightRed("ERROR"))
		end
	end
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.warn(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("W")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("W", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.lightYellow("-WARN")) .. "\n"
		else
			Col.print(display .. reason, Col.lightYellow("-WARN"))
		end
	end
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.ok(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("I")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("W", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.lightGreen("---OK")) .. "\n"
		else
			Col.print(display .. reason, Col.lightGreen("---OK"))
		end
	end
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.load(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("I")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("W", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.darkPurple("-LOAD")) .. "\n"
		else
			Col.print(display .. reason, Col.darkPurple("-LOAD"))
		end
	end
end

---@param reason string
---@param display string|nil
---@param stackTraceEnabled boolean|nil
function M.info(reason, display, stackTraceEnabled)
	if stackTraceEnabled then
		stackTrace("I")
	end
	if display == nil then
		display = cleanseName(debug.getinfo(2).source) .. "@" .. (debug.getinfo(2).name or "-")
	end
	display = Col.ifServer(display, Col.bold) .. " "

	if log then
		log("I", display, reason)
	else
		if collectMode then
			collectBuffer = collectBuffer .. Col.build(display .. reason, Col.bold("-INFO")) .. "\n"
		else
			Col.print(display .. reason, Col.bold("-INFO"))
		end
	end
end

return M
