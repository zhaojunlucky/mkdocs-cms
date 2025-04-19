import { Injectable } from '@angular/core';
import {environment} from '../../environments/environment';
import {StrUtils} from '../shared/utils/str.utils';

@Injectable({
  providedIn: 'root'
})
export class VditorUploadService {
  uploadConfig = {
    url: `${ environment.apiServer}/v1/storage`,
    withCredentials: true,
    accept: 'image/*',
    filename:(name: string) => name.replace(/\W/g, ''),
    multiple: false,
    format: (files: any, response: string) => {
      // This function handles the server's response and formats it for Vditor
      try {
        const res = JSON.parse(response);
        if (Object.keys(res.errorFiles).length > 0) {
          console.log(res.errorFiles);
        }
        let uploadedFiles = res.uploadedFiles;
        if (!environment.production) {
          let prefix = environment.apiServer.replaceAll("/api", "");
          for (const key in uploadedFiles) {
            uploadedFiles[key] = prefix + uploadedFiles[key];
          }
        }
        return JSON.stringify({
          "msg": "",
          "code": 0,
          "data": {
            "errFiles": Object.keys(res.errorFiles),
            "succMap": uploadedFiles
          }
        });
      } catch (error) {
        // @ts-ignore
        alert('Upload failed: ' + StrUtils.stringifyHTTPErr(error));
        return '';
      }
    },
    error: (msg: any) => {
      alert(StrUtils.stringifyHTTPErr(JSON.parse(msg)));
    },
  }

  constructor() { }

  getVditorOptions() {
    return {
      upload: this.uploadConfig
    }
  }
}
