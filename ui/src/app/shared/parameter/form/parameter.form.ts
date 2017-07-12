import {Component, Output, EventEmitter, ViewChild, ElementRef, Input} from '@angular/core';
import {SharedService} from '../../shared.service';
import {Parameter} from '../../../model/parameter.model';
import {ParameterEvent} from '../parameter.event.model';
import {ParameterService} from '../../../service/parameter/parameter.service';
import {Project} from '../../../model/project.model';

@Component({
    selector: 'app-parameter-form',
    templateUrl: './parameter.form.html',
    styleUrls: ['./parameter.form.scss']
})
export class ParameterFormComponent {

    parameterTypes: string[];
    newParameter: Parameter;

    @Input() project: Project;
    @Input() suggest: Array<string>;
    @ViewChild('selectType') selectType: ElementRef;

    @Output() createParameterEvent = new EventEmitter<ParameterEvent>();

    constructor(private _paramService: ParameterService, public _sharedService: SharedService) {
        this.newParameter = new Parameter();
        this.parameterTypes = this._paramService.getTypesFromCache();
        if (!this.parameterTypes) {
            this._paramService.getTypesFromAPI().subscribe(types => this.parameterTypes = types);
        }
    }

    create(): void {
        let event: ParameterEvent = new ParameterEvent('add', this.newParameter);
        if (!this.newParameter.value) {
            switch (this.newParameter.type) {
                case 'number':
                    this.newParameter.value = '0';
                    break;
                case 'boolean':
                    this.newParameter.value = 'false';
                    break;
                default:
                    this.newParameter.value = '';
            }
        }
        this.createParameterEvent.emit(event);
        this.newParameter = new Parameter();
    }

}
