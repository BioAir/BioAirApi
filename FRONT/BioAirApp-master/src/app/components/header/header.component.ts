import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit {

  
  tempHoy: any;
  constructor(private router: Router) { }

  ngOnInit(): void {
    this.tempHoy = localStorage.getItem('temp-hoy')
  }

  logout(){
    localStorage.removeItem('metadata');
    localStorage.removeItem('token');
    this.router.navigateByUrl('/inicio');
    
  }

  goToProfile(){
    this.router.navigateByUrl('/profile');
  }

  goToHome(){
    this.router.navigateByUrl('/inicio');
  }

}
