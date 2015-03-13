# goback
GoBack - image flasher for parallella, wandboard, odroid and more  
Complete information: http://mkaczanowski.com/building-arm-cluster-part-2-create-and-write-system-image-with-goback/

## How does it work?:
 1. Loads “steps” from config (StepList struct is a linked-list)
 2. Read lines from serial line
 3. If line contains “Step.Expect” or “Step.Trigger” phrase, then program will continue to the next step (Expect) or will execute onTrigger (Trigger)
 4. Program finishes when it reaches the end of the list or there is an error

Define your config in configs, edit flasher/flasher.go and run!


TODO:  
Move /config/* to a file and parse it via goback
