// getCookieValue
export const getCookieValue = (key) => {
  let result;
  let cookieKey = key + "=";
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

// removeCookie
export const removeCookie = (name) => {
  document.cookie = name + "=; expires=Thu, 01 Jan 1999 00:00:10 GMT;";
};

export const numberWithCommas = (x) => {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
};
