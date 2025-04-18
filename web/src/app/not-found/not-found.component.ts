import { Component } from '@angular/core';
import {NavComponent} from '../nav/nav.component';
import {PageTitleService} from '../services/page.title.service';

@Component({
  selector: 'app-not-found',
  imports: [
    NavComponent
  ],
  templateUrl: './not-found.component.html',
  styleUrl: './not-found.component.scss'
})
export class NotFoundComponent {
  constructor(     private pageTitleService: PageTitleService
  ) {
    this.pageTitleService.title = 'Not Found';
  }
}
