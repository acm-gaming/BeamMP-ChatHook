-- put this on the top of any to be profiled script main.lua
package.loaded["profiler/profiler"] = nil
pcall(require, "profiler/profiler")

-- Rest of script starts here
-- ...
