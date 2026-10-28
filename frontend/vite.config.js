// appType 'mpa': no single-page fallback. Under `wails dev` the web view asks Vite first for /image/N, and WebKit's image Accept
// header ends in */*, so the fallback answered with index.html and the picture could not be shown; now Vite answers 404 and Wails
// passes the request on to Go's image handler.
export default { appType: 'mpa' };
