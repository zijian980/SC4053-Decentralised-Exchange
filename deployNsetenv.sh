#!/bin/bash

npx hardhat ignition deploy ignition/modules/proj.ts --network localhost | grep ' - ' | tail -n 5 | sed 's/.*#//; s/ - /=/' | awk -F= '{print toupper($1)"="$2}' > "./webapp/backend/.env"
