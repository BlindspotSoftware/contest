# Qemu Teststep


## Parameters

### Required Parameters
* **executable:** Name of the qemu executable. It can be an absolute path or the name of a executable in $PATH.

* **firmware:** The firmware image you want to test.


### Optional Paramters
* **logfile:** The output of the running image is copied here. If left empty the output will be discarded by setting the logfile to /dev/null.

* **mem:** The amount of RAM dedicated to the qemu instance in MB.

* **nproc:** The amount of threads available to qemu. 

* **image:** A Disk Image, which can be booted by the firmware.

* **timeout:** The time intervall until the qemu instance is forcibly shut down. Example: '4m'

* **steps:** This is a list of steps, which can consist of expect or send steps. An expect steps expects a certain output from the virtual machine. A send step will send a string to qemu. Expect steps can have an additional timeout field, which is a string, like '2m'. If left empty the **timeout** Parameter is used as timeout instead. Make sure the timout you set for an expect step is shorter than the overall timeout.  A step can have both an expect as well as a send statement; this is interpreted as an expect step, which is followed by a send step if it is successful. The steps are executed in order beginning from the first entry. Each step is blocking, meaning the next step will be executed only if the previous step was successful. 