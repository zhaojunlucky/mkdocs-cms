import { Injectable } from '@angular/core';
import { CanDeactivate } from '@angular/router';
import { Observable } from 'rxjs';

// Define an interface your component will implement
export interface CanComponentDeactivate {
  canDeactivate: () => Observable<boolean> | Promise<boolean> | boolean;
}

@Injectable({
  providedIn: 'root'
})
export class CanDeactivateFormGuard implements CanDeactivate<CanComponentDeactivate> {
  canDeactivate(
    component: CanComponentDeactivate // The component being deactivated
  ): Observable<boolean> | Promise<boolean> | boolean {
    // Call the canDeactivate method on the component itself
    return component.canDeactivate ? component.canDeactivate() : true;
  }
}
