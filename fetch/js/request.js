const Headers = require('./headers');

export default class Request {
  constructor(input, {method='GET', headers={}, body=null, redirect='manual'}={}) {
    if (input instanceof Request) {
      const {
        otherURL,
        otherMethod,
        otherHeaders,
        otherRedirect,
      } = input;

      this.url = otherURL;
      this.method = otherMethod;
      this.headers = new Headers(otherHeaders);
      this.redirect = otherRedirect;
    } else {
      this.url = input;
    }

    this.method = method;
    this.headers = new Headers(headers);
    this.body = body;
    this.redirect = redirect;
  }
}