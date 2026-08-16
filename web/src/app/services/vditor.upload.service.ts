import { Injectable } from '@angular/core';
import {environment} from '../../environments/environment';
import {StrUtils} from '../shared/utils/str.utils';
import {MatDialog} from '@angular/material/dialog';
import {ConfirmDialogComponent} from '../shared/dialogs/confirm-dialog.component';

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
        return JSON.stringify({
          "msg": "",
          "code": 0,
          "data": {
            "errFiles": Object.keys(res.errorFiles),
            "succMap": res.uploadedFiles
          }
        });
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        this.showUploadError('Upload failed: ' + message);
        return '';
      }
    },
    error: (msg: any) => {
      this.showUploadError(StrUtils.stringifyHTTPErr(JSON.parse(msg)));
    },
  }

  constructor(private dialog: MatDialog) { }

  getVditorOptions() {
    return {
      upload: this.uploadConfig
    }
  }

  private showUploadError(message: string): void {
    this.dialog.open(ConfirmDialogComponent, {
      data: {
        title: 'Upload failed',
        message,
        confirmLabel: 'OK',
        cancelLabel: null,
        destructive: true,
        icon: 'error'
      }
    });
  }
}
