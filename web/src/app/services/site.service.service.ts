import { Injectable } from '@angular/core';
import {environment} from '../../environments/environment';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';
interface Version {
  version: string
}
@Injectable({
  providedIn: 'root'
})
export class SiteServiceService {

  private apiUrl = environment.apiServer;

  constructor(
    private http: HttpClient
  ) { }

  getVersion(): Observable<Version> {
    return this.http.get<Version>(`${this.apiUrl}/site/version`);
  }
}
