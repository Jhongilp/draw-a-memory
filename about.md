# stack for client side

- React
- TypeScript
- Vite

# stack for the server

- Goland

# AI stack: Google Gemini SDK for Goland

# About the app

I want to create a web app where users can create a personalized baby memory book, babe journal or keepsake journal. The app must be simple and intuitive. Basically the users will upload the photos, with that, using AI, the app will create the memory book. It must be similar to a physical book where there will be several pages as needed. The app will recognize the activities that were done and dates based on the metadata of the photos. That way it can create what was done the first month, the second and so on. Same way it can group main milestones or activities like the first visit to the pool or the zoo.  

This is the flow I am thinking:

1. User upload a group of photos
2. The app will recognize and group the photos by the date range and or activities, e.g., first visit to the zoo
3. Propose to the user a description based on the group of photos, e.g., This day we visit the zoo with your grandparents and your favorite animal was the tiger
4. With the final description, elaborate a theme to decorate the page of the book
5. Ask for approval of the theme. Once approved, save as a page in the book
6. The user could append more pages based on the new uploaded photos 

Based on this context, Act as a Senior Product Designer and React Developer. I need a web application called 'BabySteps AI Journal'.

**Core Value:** Users upload unstructured baby photos, and the app auto-magically structures them into a digital scrapbook.

**Key Features:**

1. **Smart Ingest:** A drag-and-drop zone that accepts multiple photos.
2. **AI Service: Use Gemini AI to** create the cluster based on the uploaded photos and providing the information about the date, location and suggested theme and description.
3. **The 'Magic' Draft:** For each cluster, generate a 'Page Draft'. This draft includes a suggested Title, a calculated date string (e.g., '3 Months Old'), and a suggested theme (e.g., 'Zoo', 'Cozy', 'Party') based on the 'vibe' of the group.
4. **Editor:** Allow the user to tweak the AI's suggestion (edit text, change theme) before 'Signing Off' on the page.
5. **Book View:** A main view that shows the finished book as a vertical timeline or flipbook.

**Tech Stack:** React, Tailwind CSS, Lucide-React for icons. Use soft, pastel colors and rounded aesthetics to match a baby/parenting theme.