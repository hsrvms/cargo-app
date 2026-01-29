/*
    Global authentication error handler
    Intercepts 401 responses and redirects to login
*/

(function () {
    const originalFetch = window.fetch;

    window.fetch = function (...args) {
        return originalFetch
            .apply(this, args)
            .then(async (response) => {
                if (response.status === 401) {
                    console.warn("Authentication token expired. Redirecting to login...");
                    document.cookie = "auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;";
                    window.location.href = "/login";
                    return Promise.reject(new Error("Authentication required"));
                }
                return response;
            })
            .catch((error) => {
                throw error;
            });
    };

    document.addEventListener("DOMContentLoaded", function () {
        document.body.addEventListener("htmx:responseError", function (evt) {
            if (evt.detail.xhr.status === 401) {
                console.warn("HTMX: Authentication token expired. Redirecting to login...")
                document.cookie = "auth_token=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;";
                window.location.href = "/login";
            }
        })
    })
})()
