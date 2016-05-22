import React, {Component, PropTypes} from 'react';
import {Card, CardText, CardHeader} from 'material-ui/Card';
import {
    Table,
    TableBody,
    TableHeader,
    TableHeaderColumn,
    TableRow,
    TableRowColumn
} from 'material-ui/Table';

class DataTable extends Component {
    static propTypes = {
        data: PropTypes.arrayOf(PropTypes.object).isRequired
    }

    render() {
        return (
            <Card style={{
                marginTop: "20px",
                textAlign: "left"
            }}>
                <CardHeader title={"Most Recent Data"}/>
                <CardText>
                    <Table selectable={false}>
                        <TableHeader enableSelectAll={false} displaySelectAll={false} adjustForCheckbox={false}>
                            <TableRow>
                                <TableHeaderColumn>Timestamp</TableHeaderColumn>
                                <TableHeaderColumn>Data</TableHeaderColumn>
                            </TableRow>
                        </TableHeader>
                        <TableBody displayRowCheckbox={false}>
                            {this.props.data.map((d) => {
                                let t = new Date(d.timestamp * 1000);
                                let ts = t.getHours() + ":" + t.getMinutes() + ":" + t.getSeconds() + " - " + (t.getMonth() + 1) + "/" + t.getDate() + "/" + t.getFullYear();
                                return (
                                    <TableRow key={JSON.stringify(d)}>
                                        <TableRowColumn>{ts}</TableRowColumn>
                                        <TableRowColumn>{JSON.stringify(d.data)}</TableRowColumn>
                                    </TableRow>
                                );
                            })}

                        </TableBody>
                    </Table >
                </CardText>
            </Card>
        );
    }
}
export default DataTable;
