import React, {Component, PropTypes} from 'react';
import {connect} from 'react-redux';

import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';

import TimeDifference from './TimeDifference';

class UserView extends Component {

    static propTypes = {
        user: PropTypes.shape({name: PropTypes.string.isRequired}).isRequired
    }

    render() {
        let description = (this.props.user.description === undefined
            ? ""
            : this.props.user.description);

        return (
            <div>{description == ""
                    ? (null)
                    : (
                        <div style={{
                            color: "grey"
                        }}><ReactMarkdown escapeHtml={true} source={description}/></div>
                    )
}
                <Table selectable={false}>
                    <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                        <TableRow>
                            <TableHeaderColumn>Email</TableHeaderColumn>
                            <TableHeaderColumn>Public</TableHeaderColumn>
                            <TableHeaderColumn>Role</TableHeaderColumn>
                            <TableHeaderColumn>Queried</TableHeaderColumn>
                        </TableRow>
                    </TableHeader>
                    <TableBody displayRowCheckbox={false}>
                        <TableRow>
                            <TableRowColumn>{this.props.user.email}</TableRowColumn>
                            <TableRowColumn>{this.props.user.public
                                    ? "true"
                                    : "false"}</TableRowColumn>
                            <TableRowColumn>{this.props.user.role}</TableRowColumn>
                            <TableRowColumn><TimeDifference timestamp={this.props.user.timestamp}/></TableRowColumn>
                        </TableRow>
                    </TableBody>

                </Table>
            </div>
        );
    }
}
export default UserView;
