import {GenericMap} from '../core/data.type';

export class StrUtils {
  static stringifyHTTPErr(err: object): string {
    if (!err) {
      return "Unknown error!!!"
    }


    if (err.hasOwnProperty("error")) {
      // @ts-ignore
      err = err.error
    }
    if (!err.hasOwnProperty("errorMessages")) {
      if (err.hasOwnProperty('message')) {
        // @ts-ignore
        return err['message']
      }

      return "Unknown error!!!"
    }


    let msgs = []
    if (err.hasOwnProperty('httpStatusCode')) {
      // @ts-ignore
      msgs.push(`HTTP status: ${err['httpStatusCode']}. `)
    }

    if (err.hasOwnProperty('errorMessages')) {
      // @ts-ignore
      let errorMessages = err['errorMessages']
      // @ts-ignore
      msgs.push(`Message: ${errorMessages.map(v => v['message']).join(" ")}`)
    }
    return msgs.join("")
  }

  static parseRedirectUrl(encodedUrl: string): GenericMap {
    if (!encodedUrl) {
      return {
        paths: ['/'],
        queryParams: {}
      }
    }

    let url = decodeURIComponent(encodedUrl)

    let qIndex = url.indexOf('?')
    let queryParams = {}
    if (qIndex > 0) {
      let q = url.substring(qIndex + 1)
      let args = q.split("&")
      for (let arg of args) {
        let parts = arg.split("=");
        // @ts-ignore
        queryParams[parts[0]] = parts.length > 1 ? parts[1] : ''
      }
    }
    url = url.substring(0, qIndex).replace(/^\/+|\/+$/g, '');
    let paths = url.split("/").filter(s=>s.length>0);
    if (paths.length == 0) {
      paths = ['/']
    } else {
      paths[0] = '/' + paths[0]
    }
    return {
      'paths': paths,
      'queryParams': queryParams
    }

  }
}
