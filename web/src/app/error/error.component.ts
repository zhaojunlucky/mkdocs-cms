import {Component, OnInit} from '@angular/core';
import {ActivatedRoute, RouterLink} from '@angular/router';
import {NgIf} from '@angular/common';
import {MatCardModule} from '@angular/material/card';

@Component({
  selector: 'app-error',
  imports: [
    NgIf,
    MatCardModule,
  ],
  templateUrl: './error.component.html',
  styleUrl: './error.component.scss'
})
export class ErrorComponent implements OnInit {
  error = '';
  redirect = '';
  constructor(private route: ActivatedRoute,) {
  }
  ngOnInit(): void {
    this.route.queryParamMap.subscribe(queryParams => {
      this.error = queryParams.get('error') || '';
      this.redirect = queryParams.get('redirect') || '';
    });
  }

}
