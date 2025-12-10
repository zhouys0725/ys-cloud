#!/bin/sh

# Remove .env file to force using environment variables
rm -f .env

# Start the application
exec ./main