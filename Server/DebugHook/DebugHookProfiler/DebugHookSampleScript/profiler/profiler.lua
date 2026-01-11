
--[[
package.loaded["profiler/profiler"] = nil
pcall(require, "profiler/profiler")
]]

local function split(string, delimeter, convert_into)
	local t = {}
	for str in string.gmatch(string, "([^"..delimeter.."]+)") do
		if convert_into == 1 then -- number
			table.insert(t, tonumber(str))
			
		elseif convert_into == 2 then -- bool
			if str:lower() == "false" then
				table.insert(t, false)
			elseif str:lower() == "true" then
				table.insert(t, false)
			end
			
		else -- string
			table.insert(t, str)
		end
	end
	return t
end

local function tableSize(table)
	if type(table) ~= "table" then return 0 end
	local len = 0
	for k, v in pairs(table) do
		len = len + 1
	end
	return len
end

local function filePath(string)
	local _, pos = string:find(".*/")
	if pos == nil then return nil end
	
	return string:sub(1, pos)
end

local function myPath()
	local source_path = debug.getinfo(2).source:gsub("\\", "/")
	if source_path:sub(1, 1) == '@' then return filePath(source_path:sub(2)) end
	return filePath(source_path)
end

local function scriptName()
	local script_name = split(myPath(), '/')
	return script_name[#script_name - 1]
end


function profilerRoutine()
	local readout = Util.DebugExecutionTime()
	readout.profilerRoutine = nil
	readout.onInit = nil
	if tableSize(readout) == 0 then return end
	MP.TriggerGlobalEvent("onProfilerData", scriptName(), readout)
end

local function onInit()
	MP.RegisterEvent("profilerRoutine", "profilerRoutine")
	MP.CancelEventTimer("profilerRoutine")
	MP.CreateEventTimer("profilerRoutine", 30000)
end

onInit()
