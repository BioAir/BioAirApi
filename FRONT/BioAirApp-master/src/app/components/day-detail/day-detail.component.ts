import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router'

@Component({
  selector: 'app-day-detail',
  templateUrl: './day-detail.component.html',
  styleUrls: ['./day-detail.component.css']
})
export class DayDetailComponent implements OnInit {

  day: any;
  constructor(private ruta:ActivatedRoute) {
    this.ruta.params.subscribe(params=>{
      this.day = params;
    })
   }

  ngOnInit(): void {
  }

}
