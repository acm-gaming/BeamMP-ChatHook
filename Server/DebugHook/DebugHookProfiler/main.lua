
--[[
	Format
	["script_name"] = table
		["event_name"] = table
			[n] = float (amount of executions)
			[mean] = float (avg exec time)
			[max] = float (highest)
			[min] = float (lowest)
			[stdev] = float (???)
]]
local Stats = {}

local function tableSize(table)
	if type(table) ~= "table" then return 0 end
	local len = 0
	for k, v in pairs(table) do
		len = len + 1
	end
	return len
end


local function log(string, no_chathook)
	--print(string)
	if not no_chathook then
		MP.TriggerGlobalEvent('onDebugMessage', string)
	end
end

local function limit(number)
	return string.format("%.2f", number)
end

-- min/max are swapped in Util.DebugExecutionTime(). as in max is actually min and vise versa. that gotta be a bug
function printRoutine()
	local data = ""
	local lines = 0
	for script_name, readout in pairs(Stats) do
		data = data .. '**' .. script_name .. '**\n```'
		lines = lines + 1
		for event_name, profiler in pairs(readout) do
			data = data .. string.format(
				'avg. %s ms	| min. %s ms	| max. %s ms	| %s (%s)\n',
				limit(profiler.mean), limit(profiler.max), limit(profiler.min), event_name, profiler.n
			)
			lines = lines + 1
		end
		data = data .. '```'
		
		if lines > 20 then
			log(data)
			data = ""
			lines = 0
		else
			data = data .. '\n'
			lines = lines + 1
		end
	end
end

function onProfilerData(script_name, readout)
	if script_name == nil or readout == nil then return end
	
	readout.onInit = nil
	readout.profilerRoutine = nil
	readout.printRoutine = nil
	readout.onProfilerData = nil
	if tableSize(readout) == 0 then return end
	
	Stats[script_name] = readout
end

function onInit()
	MP.RegisterEvent("onProfilerData", "onProfilerData")
	MP.RegisterEvent("printRoutine", "printRoutine")
	MP.CancelEventTimer("printRoutine")
	MP.CreateEventTimer("printRoutine", 30 * 60000)
end