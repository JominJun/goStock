export const createCookie = (cookieName, cookieValue, expirationDate) => {
  document.cookie = `${cookieName}=${cookieValue}; Expires=${expirationDate}; Secure)`;
  window.location.reload();
};

export const removeCookie = (cookieName) => {
  document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1999 00:00:10 GMT;`;
  window.location.reload();
};

export const getCookieValue = (key) => {
  let cookieKey = key + "=";
  let result = "";
  const cookieArr = document.cookie.split(";");

  for (let i = 0; i < cookieArr.length; i++) {
    if (cookieArr[i][0] === " ") {
      cookieArr[i] = cookieArr[i].substring(1);
    }

    if (cookieArr[i].indexOf(cookieKey) === 0) {
      result = cookieArr[i].slice(cookieKey.length, cookieArr[i].length);
      return result;
    }
  }
  return result;
};
