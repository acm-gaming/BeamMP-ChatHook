# DebugHook (for Advanced Users)

The DebugHook is a experimental idea of automatically sending script related events to private discord channels. This can be debug, profiling or any other data not ment for the public eye.


To make this work as intended you want to run a second ChatHook docker container and have it listen to another port.


Setup is then very easy.
1. Take a copy of /Server/ChatHook and name it DebugHook.
2. Configure as you would but ensure the port aligns to the second docker container.
3. In the main.lua also set the variable IS_DEBUG_HOOK to true
4. Save and run.


So far only a single addon makes use of this.
- Take a look at the DebugHookProfiler folder
