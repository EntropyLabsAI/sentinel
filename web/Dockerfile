# Use the official Node.js 18 Alpine image for a lightweight setup
FROM node:18-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock) to install dependencies
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application code
COPY . .

# Expose port 3000 (default React development server port)
EXPOSE 3000

# Start the React development server
CMD ["npm", "run", "dev"]
