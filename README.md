# Yates

Yates is a 9p fs that serves mutex for inter-program synchronization over network (or not).

### Start the fs
```
% yates &
```
### Create a new mutex `foo`
```
% 9p ls yates
del
new
% echo foo |9p write yates/new
% 9p ls yates
del
foo
new
%
```
### Acquire mutex
Acquiring a mutex returns a UUID that must be used to release the mutex (in order to avoid mutex stealing from an other programm).
```
% 9p read yates/foo
95fdc5ed-808a-4613-bbd6-d9249d5dc022
%
```

### Waiting for mutex
Trying to acquire already own mutex will block the acquiring programm, until mutex is released.
```
% 9p read yates/foo
...
```

### Release mutex
To release a mutex, write the associated UUID in it.
```
% echo 95fdc5ed-808a-4613-bbd6-d9249d5dc022 |9p write yates/foo
%
```
Now the mutex is released the other programm acquieres it.
```
% 9p read yates/foo
...
bc8d9966-55a9-4911-9c34-bf9f547c1b58
%
```
And so on.

### Delete mutex
```
% echo foo |9p write yates/del
% 9p ls yates
del
new
%
```
