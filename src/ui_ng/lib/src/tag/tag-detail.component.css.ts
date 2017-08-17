export const TAG_DETAIL_STYLES: string = `
.overview-section {
    background-color: white;
    padding-bottom: 36px;
    border-bottom: 1px solid #cccccc;
}

.detail-section {
    background-color: #fafafa;
    padding-left: 12px;
    padding-right: 24px;
}

.title-block {
    display: inline-block;
}

.title-wrapper {
    padding-top: 12px;
}

.tag-name {
    font-weight: 300;
    font-size: 32px;
}

.tag-timestamp {
    font-weight: 400;
    font-size: 12px;
    margin-top: 6px;
}

.rotate-90 {
    -webkit-transform: rotate(-90deg);
    /*Firefox*/
    -moz-transform: rotate(-90deg);
    /*Chrome*/
    -ms-transform: rotate(-90deg);
    /*IE9 、IE10*/
    -o-transform: rotate(-90deg);
    /*Opera*/
    transform: rotate(-90deg);
}

.arrow-back {
    cursor: pointer;
}

.arrow-block {
    border-right: 2px solid #cccccc;
    margin-right: 6px;
    display: inline-flex;
    padding: 6px 6px 6px 12px;
}

.vulnerability-block {
    margin-bottom: 12px;
}

.summary-block {
    margin-top: 24px;
    display: inline-flex;
    flex-wrap: row wrap;
}

.image-summary {
    margin-right: 36px;
    margin-left: 18px;
}

.flex-block {
    display: inline-flex;
    flex-wrap: row wrap;
    justify-content: space-around;
}

.vulnerabilities-info {
    padding-left: 24px;
}

.vulnerabilities-info .third-column {
    margin-left: 36px;
}

.vulnerabilities-info .second-column,
.vulnerabilities-info .fourth-column {
    text-align: left;
    margin-left: 6px;
}

.vulnerabilities-info .second-row {
    margin-top: 6px;
}

.detail-title {
    font-weight: 500;
    font-size: 14px;
}

.image-detail-label {
    text-align: right;
}

.image-detail-value {
    text-align: left;
    margin-left: 6px;
    font-weight: 500;
}
.tip-icon-medium {
    color: orange;
}
.tip-icon-low{
    color:yellow;
}
.second-column div, .fourth-column div, .image-detail-value div{
height: 24px;
}

`;