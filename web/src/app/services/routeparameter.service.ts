import { Injectable } from '@angular/core';
import { Router, NavigationEnd, ActivatedRoute } from '@angular/router';
import { BehaviorSubject, Observable, filter, map } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class RouteParameterService {
  private childIdSubject = new BehaviorSubject<string | null>(null);
  public childId$: Observable<string | null> = this.childIdSubject.asObservable();

  constructor(private router: Router, private activatedRoute: ActivatedRoute) {
    this.router.events.pipe(
      filter(event => event instanceof NavigationEnd),
      map(() => this.activatedRoute),
      map(route => {
        while (route.firstChild) {
          route = route.firstChild;
        }
        return route;
      }),
      filter(route => route.outlet === 'primary'),
      map(route => route.snapshot.paramMap.get('collectionName'))
    ).subscribe(childId => {
      this.childIdSubject.next(childId);
    });
  }
}
