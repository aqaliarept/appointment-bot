#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "Testing Monetra Appointment API..."

# Current date in ISO format
START_DATE=$(date -u +"%Y-%m-%dT%H:%M:%S")
# End date (2 months from now)
END_DATE=$(date -u -v+2m +"%Y-%m-%dT%H:%M:%S")

# Test first appointment type
echo -e "\n${GREEN}Testing first appointment type...${NC}"
curl -v -X POST \
  'https://outlook.office365.com/BookingsService/api/V1/bookingBusinessesc2/monetrapirkanmaarekrytointipalvelut@monetra.fi/GetStaffAvailability?app=BookingsC1' \
  -H 'Content-Type: application/json' \
  -d "{
    \"serviceId\": \"1df7f565-8337-412b-91ec-b8ffd49fe6f2\",
    \"staffIds\": [\"4f3b2516-99cd-4295-9328-afefb3b403e3\"],
    \"startDateTime\": {
        \"dateTime\": \"$START_DATE\",
        \"timeZone\": \"FLE Standard Time\"
    },
    \"endDateTime\": {
        \"dateTime\": \"$END_DATE\",
        \"timeZone\": \"FLE Standard Time\"
    }
}" | jq '.'

echo -e "\n${GREEN}Testing second appointment type...${NC}"
curl -v -X POST \
  'https://outlook.office365.com/BookingsService/api/V1/bookingBusinessesc2/monetrapirkanmaarekrytointipalvelut@monetra.fi/GetStaffAvailability?app=BookingsC1' \
  -H 'Content-Type: application/json' \
  -d "{
    \"serviceId\": \"51b3c1e4-2dc8-46ab-88e3-604cb4164c4c\",
    \"staffIds\": [\"84d3f0dd-33f9-4d2d-a741-98b86e790315\"],
    \"startDateTime\": {
        \"dateTime\": \"$START_DATE\",
        \"timeZone\": \"FLE Standard Time\"
    },
    \"endDateTime\": {
        \"dateTime\": \"$END_DATE\",
        \"timeZone\": \"FLE Standard Time\"
    }
}" | jq '.' 