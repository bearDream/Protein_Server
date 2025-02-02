# Protein Projects

This project contains the backend of the following projects:
* PROFASA
* PS-GO
* ProteinFlow
Techniques used include:
* Gin (https://github.com/gin-gonic/gin)
* Gorm (https://github.com/go-gorm/gorm)
* MySQL


## Available Scripts

In the project directory, you can run:


### `yarn start`

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in the browser.

### `yarn test`

Launches the test runner in the interactive watch mode.\
See the section about [running tests](https://facebook.github.io/create-react-app/docs/running-tests) for more information.

### `yarn build`

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

The build is minified and the filenames include the hashes.\
Your app is ready to be deployed!

See the section about [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.

## Upload

### Login Filezilla
Login to the `ym5@csgate.ucc.ie`.

Upload the `build` folder to the `public_html/aipo`, you can delete all the files in the `public_html/aipo` before upload.

### Login Terminal

`ssh ym5@csgate.ucc.ie`

`scp -r /users/postgrads/ym5/public_html/aipo ym5@csweb.ucc.ie:/www/docs/multimedia/Public/training`
