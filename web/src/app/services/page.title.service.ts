import { Injectable } from '@angular/core';
import {Title} from '@angular/platform-browser';

@Injectable({
  providedIn: 'root'
})
export class PageTitleService {
  base = 'MkDocs CMS'
  _title = this.base;

  constructor(private bodyTitle: Title) {

  }

  get title(): string {
    return this._title;
  }

  set title(title: string) {
    this._title = `${title} | ${this.base}`;
    this.bodyTitle.setTitle(this._title);
  }
}
