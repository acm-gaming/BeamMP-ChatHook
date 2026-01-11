# DebugHook (Profiler Addon - Advanced Users)

This addon allows measuring all lua states event runtime performance.
![img](https://github.com/OfficialLambdax/BeamMP-ChatHook/blob/main/img/debughook_profiler.jpg?raw=true)


In order to set this up
1. Take a copy of this folder and put it into your Resources/Server folder
2. Done


This script alone will only care to collect the information to then send it to the DebugHook. Todo that it must be able to access each state. The way this have been solved is by simply requiring a single file in each lua state.


Beaware that this "may only" work when a lua state is booted from a single entrypoint, such as a main.lua.
- ScriptFolder/main.lua
- ScriptFolder/libs/anyother.lua


and not when
- ScriptFolder/main.lua
- ScriptFolder/anyother.lua

(atleast i myself had different results when a state has multiple entrypoints).


If this is the case then simply open the single entrypoint file (eg main.lua) and on the very top copy 2 lines.
```lua
package.loaded["profiler/profiler"] = nil
pcall(require, "profiler/profiler")
```

Then copy the "profiler" folder fom the DebugHookSampleScript folder into the same folder your main.lua file is in. Save and done.

The profiler will send an event to the DebugHook every 30 Minutes.
